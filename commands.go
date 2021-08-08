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

// Connect
func (c *client) Connect() error {
	var err error
	var u url.URL

	if c.config.Connection.secure {
		u = url.URL{Scheme: "wss", Host: twitchWSSHost}
	} else {
		u = url.URL{Scheme: "ws", Host: twitchWSHost}
	}

	// Make sure the connection is not already open before connecting
	if c.conn != nil {
		err = c.disconnect()
	}
	if err != nil {
		return err
	}

	c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	var wg = &sync.WaitGroup{}
	if c.config.Connection.sync {
		wg.Add(1)
	}
	go c.readMessages(wg)

	err = c.sendConnectSequence()
	if err != nil {
		return err
	}

	for _, channel := range c.config.Channels {
		err = c.Join(channel)
		if err != nil {
			return err
		}
	}

	if c.config.Connection.sync {
		wg.Wait()
	}

	return nil
}

// Disconnect sends a close message to the server and lets the server close the connection.
// If an error occurs telling the server to close, the connection is closed without waiting.
func (c *client) Disconnect() error {
	c.userDisconnect.Notify() // let all relevant goroutines know the user called Disconnect
	return c.disconnect()
}

// disconnect is a helper function for Disconnect that allows for server disconnect calls
// without notifying goroutines of the disconnect.
func (c *client) disconnect() error {
	c.wMutex.Lock()
	err := c.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.wMutex.Unlock()
	if err != nil {
		return c.conn.Close()
	}
	return nil
}

// Done sets the callback function to be called when a client is done (typically due to a fatal error).
func (c *client) Done(cb func()) {
	c.done = cb
}

// Failure returns the fatal error that caused client.done to be called if there was an error.
func (c *client) Failure() error {
	return c.fatal
}

// Join joins channel.
func (c *client) Join(channel string) error {
	if !strings.HasPrefix(channel, "#") {
		channel = "#" + channel
	}

	return c.send("JOIN " + channel)
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
	ircData := parseIRCMessage(rawMessage)

	// TODO: parseRawMessage for each else below, and call the UNSET handler
	switch ircData.Prefix {
	case "tmi.twitch.tv":
		if h, ok := tmiTwitchTvHandlers(ircData.Command); ok {
			if h != nil {
				h(c, ircData)
			}
		} else {
			if c.onError != nil {
				c.onError(errors.New("could not handle message with tmi.twitch.tv prefix:\n" + rawMessage))
			}
		}
	case "jtv":
		if h, ok := jtvHandlers(ircData.Command); ok {
			if h != nil {
				h(c, ircData)
			}
		} else {
			if c.onError != nil {
				c.onError(errors.New("could not handle message with jtv prefix:\n" + rawMessage))
			}
		}
	default:
		if h, ok := otherHandlers(ircData.Command); ok {
			if h != nil {
				h(c, ircData)
			}
		} else {
			if c.onError != nil {
				c.onError(errors.New("could not handle message with { " + ircData.Prefix + " } as prefix:\n" + rawMessage))
			}
		}
	}
}

func (c *client) readMessages(wg *sync.WaitGroup) {
	if c.config.Connection.sync {
		defer wg.Done()
	}
	for {
		c.rMutex.Lock()
		_, received, err := c.conn.ReadMessage()
		c.rMutex.Unlock()
		if err != nil {
			return
		}
		select {
		case c.rcvdMsg <- true:
		default:
		}

		data := strings.Split(string(received), "\r\n")
		for _, rawMessage := range data {
			if len(rawMessage) > 0 {
				c.handleMessage(rawMessage)
			}
		}
	}
}

// TODO: change back to private, and use it
func (c *client) Reconnect() error {
	var maxAttempts = c.config.Connection.maxReconnectAttempts
	if maxAttempts == 0 {
		return errors.New("tmi.client.reconnect(): max attempts was 0")
	}
	var maxInterval = time.Duration(c.config.Connection.maxReconnectInterval)
	if maxInterval < 0 {
		return errors.New("tmi.client.reconnect(): max interval was negative")
	}

	var err = c.Connect()
	for i := 1; err != nil; i++ {
		if maxAttempts >= 0 && i >= maxAttempts {
			return errors.New("tmi.client.reconnect(): max attempts to reconnect reached")
		}
		sleepDuration := time.Duration(math.Pow(2, float64(i)))
		if sleepDuration > maxInterval {
			sleepDuration = maxInterval
		}
		time.Sleep(sleepDuration)
		err = c.Connect()
	}
	return nil
}

func (c *client) send(message string) error {
	c.wMutex.Lock()
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(message))
	c.wMutex.Unlock()
	return err
}

func (c *client) sendConnectSequence() error {
	var err error
	err = c.send("PASS " + c.config.Identity.password)
	if err != nil {
		return err
	}
	err = c.send("NICK " + c.config.Identity.username)
	if err != nil {
		return err
	}
	err = c.send("CAP REQ :twitch.tv/tags twitch.tv/commands twitch.tv/membership")
	if err != nil {
		return err
	}
	return err
}
