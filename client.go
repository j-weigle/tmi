package tmi

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client interface {
	// Connect connects to irc-ws.chat.twitch.tv and attempts to reconnect on connection errors.
	Connect() error

	// Disconnect closes the connection to the server, and does not attempt to reconnect.
	Disconnect()

	// Join joins channel.
	Join(channel string) error

	// On sets the callback function to cb for the MessageType mt.
	On(mt MessageType, cb func(Message))

	// Done sets the callback function for when a client is done to cb. Useful for running a client in a goroutine.
	OnDone(cb func(fatal error))

	// OnErr sets the callback function for general error messages to cb.
	OnErr(cb func(error))

	// Part leaves channel.
	Part(channel string) error

	// Say sends a PRIVMSG message in channel.
	Say(channel string, message string) error

	// UpdatePassword updates the password the client uses for authentication.
	UpdatePassword(string)
}

type client struct {
	config           *clientConfig
	conn             *websocket.Conn
	closeErr         connCloseErr
	closeErrCb       func(error)                   // internal callback for goroutines on connection events.
	done             func(error)                   // callback function for fatal errors.
	handlers         map[MessageType]func(Message) // callback functions for each MessageType.
	inbound          chan string                   // for sending inbound messages to the handlers, acts as a buffer.
	notifDisconnect  notifier                      // used for disconnect call notifications
	onError          func(error)                   // callback function for non-fatal errors.
	outbound         chan string                   // for sending outbound messages to the writer.
	rcvdMsg          chan bool                     // when conn reads, notifies ping loop.
	rcvdPong         chan bool                     // when pong received, notifies ping loop.
	reconnectCounter int                           // for keeping track of reconnect attempts before a successful attempt.
}

type connCloseErr struct {
	mutex sync.Mutex
	err   error
}

func (t *connCloseErr) update(err error) {
	var override = err == ErrDisconnectCalled || err == ErrLoginFailure
	t.mutex.Lock()
	if t.err == nil {
		t.err = err
	} else if override {
		if t.err != ErrDisconnectCalled && t.err != ErrLoginFailure {
			t.err = err
		}
	}
	t.mutex.Unlock()
}

func (t *connCloseErr) reset() {
	t.mutex.Lock()
	t.err = nil
	t.mutex.Unlock()
}

func (t *connCloseErr) get() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.err
}

// notifier's reset() and notify() methods are used in combination to notify multiple goroutines to close.
// call reset() before spawning goroutines
// call notify() in any goroutines to signal one another by listening to the notifier's channel ch
type notifier struct {
	mutex sync.Mutex
	once  *sync.Once
	ch    chan struct{}
}

// notify uses the notifier and makes it unusable until reset.
func (n *notifier) notify() {
	n.mutex.Lock()
	n.once.Do(func() {
		if n.ch != nil {
			close(n.ch)
		}
	})
	n.mutex.Unlock()
}

// reset sets the notifier to be ready to be used.
func (n *notifier) reset() {
	n.mutex.Lock()
	n.once = &sync.Once{}
	n.ch = make(chan struct{})
	n.mutex.Unlock()
}

// NewClient returns a new client using the provided config.
func NewClient(c *clientConfig) Client {
	var config = c.deepCopy()
	return &client{
		config:   config,
		handlers: make(map[MessageType]func(Message)),
		inbound:  make(chan string, 512),
		outbound: make(chan string, 512),
		rcvdMsg:  make(chan bool),
	}
}

func (c *client) callDone(err error) {
	if c.done != nil {
		c.done(err)
	}
}

func (c *client) connect(u url.URL) error {
	var err error
	select { // Check disconnect has not been called before bothering to reconnect.
	case <-c.notifDisconnect.ch:
		return ErrDisconnectCalled
	default:
	}

	// Make sure the connection is not already open before connecting.
	if c.conn != nil {
		c.disconnect()
	}

	// Establish a connection to the URL defined by u.
	if c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil); err != nil {
		return err
	}

	// Waitgroup and context for goroutine control.
	var wg = &sync.WaitGroup{}
	var ctx, cancelFunc = context.WithCancel(context.Background())

	c.closeErr.reset()
	// Let goroutines have a callback to signal one another to return using context's CancelFunc.
	c.closeErrCb = func(err error) {
		cancelFunc()
		c.closeErr.update(err)
	}

	// Begin reading from c.conn in separate goroutine.
	c.spawnReader(ctx, wg)

	// Send NICK, PASS, and CAP REQ.
	// Sends in this goroutine before starting writer to prevent write conflicts.
	c.sendConnectSequence()

	// Begin writing to c.conn in separate goroutine.
	c.spawnWriter(ctx, wg)

	// Start the pinger in a separate goroutine.
	// It will ping c.conn after it hasn't received a message for c.config.Pinger.wait.
	c.spawnPinger(ctx, wg)

	// TODO: c.Join(c.config.Channels)

	// Block and wait for a Disconnect() call or a connection error.
	c.readInbound(ctx)

	// Make sure reader, writer, and pinger have finished.
	wg.Wait()

	// Double check connection was closed.
	c.disconnect()

	err = c.closeErr.get()

	return err
}

// disconnect sends a close message to the server and then closes the connection.
func (c *client) disconnect() {
	c.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.conn.Close()
}

func (c *client) handleMessage(rawMessage string) {
	var ircData = parseIRCMessage(rawMessage)

	// TODO: parseRawMessage for each else below, and call the UNSET handler
	switch ircData.Prefix {
	case "tmi.twitch.tv":
		if h, ok := tmiTwitchTvHandlers(ircData.Command); ok {
			if h != nil {
				h(c, ircData)
			}
		} else {
			c.warnUser(errors.New("could not handle message with tmi.twitch.tv prefix:\n" + rawMessage))
		}
	case "jtv":
		if h, ok := jtvHandlers(ircData.Command); ok {
			if h != nil {
				h(c, ircData)
			}
		} else {
			c.warnUser(errors.New("could not handle message with jtv prefix:\n" + rawMessage))
		}
	default:
		if h, ok := otherHandlers(ircData.Command); ok {
			if h != nil {
				h(c, ircData)
			}
		} else {
			c.warnUser(errors.New("could not handle message with { " + ircData.Prefix + " } prefix:\n" + rawMessage))
		}
	}
}

func (c *client) readInbound(ctx context.Context) {
	for {
		select {
		case <-c.notifDisconnect.ch:
			c.closeErrCb(ErrDisconnectCalled)
			return
		case <-ctx.Done():
			return
		case rawMessage := <-c.inbound:
			c.handleMessage(rawMessage)
		}
	}
}

func (c *client) send(message string) error {
	select {
	case c.outbound <- message:
		return nil
	default:
		return errors.New("message not delivered to outbound channel")
	}
}

func (c *client) sendConnectSequence() (err error) {
	var message string
	message = "PASS " + c.config.Identity.password
	err = c.conn.WriteMessage(websocket.TextMessage, []byte(message+"\r\n"))
	if err != nil {
		return
	}
	message = "NICK " + c.config.Identity.username
	err = c.conn.WriteMessage(websocket.TextMessage, []byte(message+"\r\n"))
	if err != nil {
		return
	}
	message = "CAP REQ :twitch.tv/tags twitch.tv/commands twitch.tv/membership"
	err = c.conn.WriteMessage(websocket.TextMessage, []byte(message+"\r\n"))
	return
}

func (c *client) spawnPinger(ctx context.Context, wg *sync.WaitGroup) {
	// recreate each time so that there isn't a pong sitting in the channel on reconnects
	c.rcvdPong = make(chan bool, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return

			case <-c.rcvdMsg:
				continue

			case <-time.After(c.config.Pinger.wait):
				c.send("PING :tmi.twitch.tv")

				select {
				case <-c.rcvdPong:
					continue

				case <-time.After(c.config.Pinger.timeout):
					c.closeErrCb(errReconnect)
					return
				}
			}
		}
	}()
}

func (c *client) spawnReader(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select { // don't block, but check for signal to be done
			case <-ctx.Done():
				return
			default:
			}
			_, received, err := c.conn.ReadMessage()
			if err != nil {
				c.closeErrCb(errReconnect)
				return
			}
			data := strings.Split(string(received), "\r\n")
			for _, rawMessage := range data {
				if len(rawMessage) > 0 {
					select { // notify pinger to reset its wait timer for received messages, if it's listening
					case c.rcvdMsg <- true:
					default:
					}
					c.inbound <- rawMessage
				}
			}
		}
	}()
}

func (c *client) spawnWriter(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer c.disconnect()

		for {
			select {
			case <-ctx.Done():
				return

			case message := <-c.outbound:
				errWriting := c.conn.WriteMessage(websocket.TextMessage, []byte(message+"\r\n"))
				if errWriting != nil {
					c.outbound <- message // store for after reconnect

					c.closeErrCb(errReconnect)
					return
				}
			}
		}
	}()
}

func (c *client) warnUser(err error) {
	if c.onError != nil {
		c.onError(err)
	}
}
