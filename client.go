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
	Join(channels ...string) error

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
	channels         map[string]bool
	channelsMutex    sync.RWMutex
	config           *clientConfig
	conn             *websocket.Conn
	connected        atomicBool
	done             func(error)                   // callback function for fatal errors.
	handlers         map[MessageType]func(Message) // callback functions for each MessageType.
	inbound          chan string                   // for sending inbound messages to the handlers, acts as a buffer.
	joinQMutex       sync.Mutex                    // join queue mutex, prevent exceeding join rate limit
	notifDisconnect  notifier                      // used for disconnect call notifications
	onError          func(error)                   // callback function for non-fatal errors.
	outbound         chan string                   // for sending outbound messages to the writer.
	rcvdMsg          chan struct{}                 // when conn reads, notifies ping loop.
	rcvdPong         chan struct{}                 // when pong received, notifies ping loop.
	reconnectCounter int                           // for keeping track of reconnect attempts before a successful attempt.
}

// NewClient returns a new client using the provided config.
func NewClient(c *clientConfig) Client {
	config := *c
	return &client{
		channels: make(map[string]bool),
		config:   &config,
		handlers: make(map[MessageType]func(Message)),
		inbound:  make(chan string, 512),
		outbound: make(chan string, 512),
		rcvdMsg:  make(chan struct{}),
	}
}

func (c *client) callDone(err error) {
	if c.done != nil {
		c.done(err)
	}
}

func (c *client) callMessageHandler(mt MessageType, message Message) {
	if h, ok := c.handlers[mt]; ok {
		if h != nil {
			h(message)
		} else {
			c.warnUser(errors.New("message handler for type " + mt.String() + "was set, but value was nil"))
		}
	}
}

func (c *client) connect(u url.URL) error {
	var err error
	select { // Check disconnect has not been called before bothering to reconnect.
	case <-c.notifDisconnect.ch:
		return ErrDisconnectCalled
	default:
	}

	// Establish a connection to the URL defined by u.
	if c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil); err != nil {
		return err
	}

	// Waitgroup and context for goroutine control.
	var wg = &sync.WaitGroup{}
	var ctx, cancelFunc = context.WithCancel(context.Background())

	var closeErr = &connCloseErr{}
	// Let goroutines have a callback to signal one another to return using context's CancelFunc.
	var closeErrCb = func(errReason error) {
		cancelFunc()
		closeErr.update(errReason)
	}

	// Begin reading from c.conn in separate goroutine.
	c.spawnReader(ctx, wg, closeErrCb)

	// Send NICK, PASS, and CAP REQ.
	// Sends in this goroutine before starting writer to prevent write conflicts.
	c.sendConnectSequence()

	// Begin writing to c.conn in separate goroutine.
	c.spawnWriter(ctx, wg, closeErrCb)

	// Start the pinger in a separate goroutine.
	// It will ping c.conn after it hasn't received a message for c.config.Pinger.wait.
	c.spawnPinger(ctx, wg, closeErrCb)

	// Block and wait for a Disconnect() call or a connection error.
	c.readInbound(ctx, closeErrCb)

	c.connected.set(false)

	// Make sure reader, writer, and pinger have finished.
	wg.Wait()

	return closeErr.err
}

// disconnect sends a close message to the server and then closes the connection.
func (c *client) disconnect() {
	c.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.conn.Close()
}

func (c *client) handleMessage(rawMessage string) error {
	var ircData, errParseIRC = parseIRCMessage(rawMessage)
	var parseUnset = func() error {
		var unsetMessage, err = parseUnsetMessage(ircData)
		if err != nil {
			c.warnUser(err)
		}
		c.callMessageHandler(UNSET, unsetMessage)
		return nil
	}
	if errParseIRC != nil {
		return parseUnset()
	}

	switch ircData.Prefix {
	case "tmi.twitch.tv":
		if h, ok := tmiTwitchTvHandlers(ircData.Command); ok {
			if h != nil {
				return h(c, ircData)
			}
		} else {
			c.warnUser(errors.New("unrecognized message with tmi.twitch.tv prefix:\n" + rawMessage))
			return parseUnset()
		}
	case "jtv":
		if h, ok := jtvHandlers(ircData.Command); ok {
			if h != nil {
				return h(c, ircData)
			}
		} else {
			c.warnUser(errors.New("unrecognized message with jtv prefix:\n" + rawMessage))
			return parseUnset()
		}
	default:
		if h, ok := otherHandlers(ircData.Command); ok {
			if h != nil {
				return h(c, ircData)
			}
		} else {
			c.warnUser(errors.New("unrecognized message with { " + ircData.Prefix + " } prefix:\n" + rawMessage))
		}
	}

	return parseUnset()
}

// locks join queue and begins sending joins on a interval
func (c *client) joinChannels(channels []string) {
	if channels == nil || len(channels) < 1 {
		return
	}

	c.joinQMutex.Lock() // prevent exceeding join limit
	defer c.joinQMutex.Unlock()
	// join limit is 20 attempts per 10 seconds per user,
	// 0.6s interval allows a 0.1s grace period to be safe
	const interval = time.Millisecond * 600

	for _, ch := range channels {
		if c.connected.get() {
			if c.send("JOIN "+ch) == nil {
				c.channelsMutex.Lock()
				c.channels[ch] = true
				c.channelsMutex.Unlock()
				time.Sleep(interval)
			}
		}
	}
}

func (c *client) onConnectedJoins() {
	var channels = []string{}
	c.channelsMutex.Lock()
	for channel := range c.channels {
		c.channels[channel] = false
		channels = append(channels, channel)
	}
	c.channelsMutex.Unlock()
	c.joinChannels(channels)
}

func (c *client) readInbound(ctx context.Context, closeErrCb func(error)) {
	for {
		select {
		case <-c.notifDisconnect.ch:
			closeErrCb(ErrDisconnectCalled)
			return
		case <-ctx.Done():
			return
		case rawMessage := <-c.inbound:
			if err := c.handleMessage(rawMessage); err != nil {
				closeErrCb(err)
				return
			}
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

func (c *client) spawnPinger(ctx context.Context, wg *sync.WaitGroup, closeErrCb func(error)) {
	// recreate each time so that there isn't a pong sitting in the channel on reconnects
	c.rcvdPong = make(chan struct{}, 1)

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
					closeErrCb(errReconnect)
					return
				}
			}
		}
	}()
}

func (c *client) spawnReader(ctx context.Context, wg *sync.WaitGroup, closeErrCb func(error)) {
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
				closeErrCb(errReconnect)
				return
			}
			data := strings.Split(string(received), "\r\n")
			for _, rawMessage := range data {
				if len(rawMessage) > 0 {
					select { // notify pinger to reset its wait timer for received messages, if it's listening
					case c.rcvdMsg <- struct{}{}:
					default:
					}
					c.inbound <- rawMessage
				}
			}
		}
	}()
}

func (c *client) spawnWriter(ctx context.Context, wg *sync.WaitGroup, closeErrCb func(error)) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer c.disconnect()

		for {
			select {
			case <-ctx.Done():
				return

			case message := <-c.outbound:
				err := c.conn.WriteMessage(websocket.TextMessage, []byte(message+"\r\n"))
				if err != nil {
					c.outbound <- message // store for after reconnect

					closeErrCb(errReconnect)
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
