package tmi

import (
	"errors"
	"math"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	twitchWSSHost = "irc-ws.chat.twitch.tv:443"
	twitchWSHost  = "irc-ws.chat.twitch.tv:80"
)

var (
	errReconnectNotification  = errors.New("reconnect")
	errDisconnectNotification = errors.New("disconnect")
	errFatalNotification      = errors.New("fatal")
)

// Connect
func (c *client) Connect() error {
	var err error
	var u url.URL

	if c.config.Connection.secure {
		u = url.URL{Scheme: "wss", Host: twitchWSSHost}
	} else {
		u = url.URL{Scheme: "ws", Host: twitchWSHost}
	}

	var maxReconnectAttempts = c.config.Connection.maxReconnectAttempts
	var maxReconnectInterval = time.Duration(c.config.Connection.maxReconnectInterval)

	for {
		err = c.connect(u)

		switch err {
		case errReconnectNotification:
			var sleepDuration time.Duration
			var overflowPoint = 64 // technically 63, but using i - 1

			if maxReconnectAttempts == 0 {
				err = errors.New("max reconnect attempts was 0")
				c.callDone(err)
				return err
			}

			i := c.reconnectCounter
			c.reconnectCounter++

			if c.reconnectCounter < 0 { // in case of overflow, reset to overflow point in order to maintain max interval
				c.reconnectCounter = overflowPoint
			}

			if maxReconnectAttempts >= 0 && i >= maxReconnectAttempts {
				err = errors.New("max attempts to reconnect reached")
				c.callDone(err)
				return err
			}

			if i == 0 {
				sleepDuration = 0
			} else if i > 0 && i < overflowPoint {
				// i - 1 to compensate for initial reconnect attempt being 0
				sleepDuration = time.Duration(math.Pow(2, float64(i-1)))
			} else {
				sleepDuration = maxReconnectInterval
			}

			if sleepDuration > maxReconnectInterval {
				sleepDuration = maxReconnectInterval
			}

			time.Sleep(sleepDuration)
			continue
		default:
			return err
		}
	}
}

func (c *client) connect(u url.URL) error {
	var err error
	// Make sure the connection is not already open before connecting.
	if c.conn != nil {
		err = c.disconnect()
		if err != nil {
			return err
		}
	}

	// Reset the notifiers for disconnects, fatal errors, and reconnects.
	c.notifDisconnect.Reset()
	c.notifFatal.Reset()
	c.notifReconnect.Reset()

	// make sure pingerDone channel is allocated, then closed, so if pinger is never spawned, it doesn't block
	c.notifPingerDone.Reset()
	c.notifPingerDone.Notify()

	// Establish a connection to the URL defined by u.
	if c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil); err != nil {
		return err
	}

	var wg = &sync.WaitGroup{}
	wg.Add(1)
	go c.spawnReader(wg)
	wg.Add(1)
	go c.spawnWriter(wg)

	c.sendConnectSequence()

	// TODO: c.Join(c.config.Channels)

	err = c.connectionManager()
	c.notifReconnect.Notify()

	// make sure reader and writer have finished
	wg.Wait()

	// writer is done, so disconnect safe to call
	c.disconnect()

	// make sure pinger has finished
	<-c.notifPingerDone.ch

	return err
}

// Disconnect closes the connection to the server, and does not attempt to reconnect
func (c *client) Disconnect() {
	c.notifDisconnect.Notify() // let all relevant goroutines know the user called Disconnect
}

// disconnect sends a close message to the server and lets the server close the connection.
// If an error occurs telling the server to close, the connection is closed without waiting.
// Should only called from writer or before writer is spawned due to websocket.Conn.WriteMessage call.
func (c *client) disconnect() error {
	err := c.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return c.conn.Close()
	}
	return nil
}

// Done sets the callback function to be called when a client is done (typically due to a fatal error).
func (c *client) Done(cb func(fatal error)) {
	c.done = cb
}

// Join joins channel.
func (c *client) Join(channel string) error {
	if channel == "" {
		return errors.New("channel was empty string")
	}
	channel = strings.ToLower(channel)
	if !strings.HasPrefix(channel, "#") {
		channel = "#" + channel
	}

	return c.send("JOIN " + channel)
}

// TODO: handle joins without breaking Twitch JOIN limits
func (c *client) J_oin(channels []string) error {
	if channels == nil || len(channels) < 1 {
		return errors.New("channels was empty or nil")
	}
	for i, channel := range channels {
		channels[i] = strings.ToLower(channel)
		if !strings.HasPrefix(channel, "#") {
			channels[i] = "#" + channel
		}
	}

	return nil
}

// On sets the callback function to cb for the MessageType mt.
func (c *client) On(mt MessageType, cb func(Message)) {
	c.handlers[mt] = cb
}

// OnErr sets the callback function for general error messages to cb.
func (c *client) OnErr(cb func(error)) {
	c.onError = cb
}

// Part leaves channel.
func (c *client) Part(channel string) error {
	if !strings.HasPrefix(channel, "#") {
		channel = "#" + channel
	}

	return c.send("PART " + channel)
}

// Say sends a PRIVMSG message in channel.
func (c *client) Say(channel string, message string) error {
	if strings.HasPrefix(c.config.Identity.username, "justinfan") {
		return errors.New("cannot send messages as an anonymous user")
	}
	if len(message) >= 500 {
		return errors.New("twitch chat's message length limit is 500 characters")
	}

	if !strings.HasPrefix(channel, "#") {
		channel = "#" + channel
	}
	return c.send("PRIVMSG " + channel + " :" + message)
}

// UpdatePassword updates the password the client uses for authentication.
func (c *client) UpdatePassword(password string) {
	c.config.Identity.SetPassword(password)
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

func (c *client) spawnReader(wg *sync.WaitGroup) {
	defer c.notifReconnect.Notify()
	defer wg.Done()

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
}

func (c *client) spawnWriter(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-c.notifFatal.ch:
			c.disconnect()
			return

		case <-c.notifReconnect.ch:
			c.disconnect()
			return

		case <-c.notifDisconnect.ch:
			c.disconnect()
			return

		case message := <-c.outbound:
			err := c.conn.WriteMessage(websocket.TextMessage, []byte(message+"\r\n"))
			if err != nil {
				c.outbound <- message // store for after reconnect

				c.notifReconnect.Notify()
				return
			}
		}
	}
}

func (c *client) connectionManager() error {
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

func (c *client) sendConnectSequence() {
	c.send("PASS " + c.config.Identity.password)
	c.send("NICK " + c.config.Identity.username)
	c.send("CAP REQ :twitch.tv/tags twitch.tv/commands twitch.tv/membership")
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
