package tmi

import (
	"errors"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client interface {
	// Connect
	Connect() error

	// Disconnect sends a close message to the server and lets the server close the connection.
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
	done             func(error)                   // callback function for fatal errors.
	handlers         map[MessageType]func(Message) // callback functions for each MessageType.
	inbound          chan string                   // for sending inbound messages to the handlers, acts as a buffer.
	notifFatal       notifier                      // notification of fatal error.
	notifDisconnect  notifier                      // notification of user manual disconnect.
	notifReconnect   notifier                      // notification of reconnect.
	onError          func(error)                   // callback function for non-fatal errors.
	outbound         chan string                   // for sending outbound messages to the writer.
	rcvdMsg          chan bool                     // when conn reads, notifies ping loop.
	rcvdPong         chan bool                     // when pong received, notifies ping loop.
	reconnectCounter int                           // for keeping track of reconnect attempts before a successful attempt.
	wg               *sync.WaitGroup               // for waiting on all goroutines to finish.
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
		close(n.ch)
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
	var handlers = make(map[MessageType]func(Message))
	return &client{
		config:   config,
		handlers: handlers,
		inbound:  make(chan string, 512),
		outbound: make(chan string, 512),
		rcvdMsg:  make(chan bool),
	}
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

func (c *client) spawnReader() {
	c.wg.Add(1)
	go func() {
		defer c.notifReconnect.notify()
		defer c.wg.Done()

		for {
			_, received, err := c.conn.ReadMessage()
			if err != nil {
				return
			}
			data := strings.Split(string(received), "\r\n")
			for _, rawMessage := range data {
				if len(rawMessage) > 0 {
					select { // notify pinger to reset its wait timer for received messages, but don't block.
					case c.rcvdMsg <- true:
					default:
					}
					c.inbound <- rawMessage
				}
			}
		}
	}()
}

func (c *client) spawnWriter() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		for {
			select {
			case <-c.notifDisconnect.ch:
				return
			case <-c.notifFatal.ch:
				return
			case <-c.notifReconnect.ch:
				return

			case message := <-c.outbound:
				err := c.conn.WriteMessage(websocket.TextMessage, []byte(message+"\r\n"))
				if err != nil {
					c.outbound <- message // store for after reconnect

					c.notifReconnect.notify()
					return
				}
			}
		}
	}()
}

func (c *client) readInbound() error {
	for {
		select {
		case rawMessage := <-c.inbound:
			c.handleMessage(rawMessage)

		case <-c.notifReconnect.ch:
			return errReconnectNotification

		case <-c.notifDisconnect.ch:
			return errDisconnectNotification

		case <-c.notifFatal.ch:
			return errFatalNotification
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

func (c *client) warnUser(err error) {
	if c.onError != nil {
		c.onError(err)
	}
}

func (c *client) callDone(err error) {
	if c.done != nil {
		c.done(err)
	}
}

func (c *client) connect(u url.URL) error {
	var err error
	// Make sure the connection is not already open before connecting.
	if c.conn != nil {
		c.disconnect()
	}

	// Establish a connection to the URL defined by u.
	if c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil); err != nil {
		return err
	}

	//TODO: possibly replace notifiers with context
	// Reset the notifiers for disconnects, fatal errors, and reconnects.
	c.notifDisconnect.reset()
	c.notifFatal.reset()
	c.notifReconnect.reset()

	// Reset client's WaitGroup wg.
	c.wg = &sync.WaitGroup{}

	// Begin reading from c.conn.
	c.spawnReader()

	// Send NICK, PASS, and CAP REQ.
	// Sends in this goroutine before starting writer to prevent write conflicts.
	c.sendConnectSequence()

	// Begin writing to c.conn.
	c.spawnWriter()

	// TODO: c.Join(c.config.Channels)

	err = c.readInbound()
	c.notifReconnect.notify()

	// make sure reader, writer, and pinger have finished
	c.wg.Wait()

	c.disconnect()

	return err
}

func (c *client) spawnPinger() {
	// recreate each time so that there isn't a pong sitting in the channel on reconnects
	c.rcvdPong = make(chan bool, 1)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		for {
			select {
			case <-c.notifFatal.ch:
				return
			case <-c.notifReconnect.ch:
				return
			case <-c.notifDisconnect.ch:
				return

			case <-c.rcvdMsg:
				continue

			case <-time.After(c.config.Pinger.wait):
				c.send("PING :tmi.twitch.tv")

				select {
				case <-c.rcvdPong:
					continue

				case <-time.After(c.config.Pinger.timeout):
					c.notifReconnect.notify()
					return
				}
			}
		}
	}()
}
