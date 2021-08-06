package tmi

import (
	"errors"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	twitchWSSHost = "irc-ws.chat.twitch.tv:443"
	twitchWSHost  = "irc-ws.chat.twitch.tv:80"
)

// CloseConnection closes the connection using websocket.Conn.Close()
// It does not send a close message or wait to receive one.
func (c *client) CloseConnection() error {
	return c.conn.Close()
}

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
		err = c.Disconnect()
		if err != nil {
			c.CloseConnection()
		}
	}

	c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	go c.readMessages()

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

	return nil
}

// Disconnect sends a close message to the server and lets the server close the connection
func (c *client) Disconnect() error {
	c.wMutex.Lock()
	err := c.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.wMutex.Unlock()
	return err
}

func (c *client) Done(cb func()) {
	c.done = cb
}

func (c *client) Err() error {
	var err = c.err
	return err
}

// Join joins channel
func (c *client) Join(channel string) error {
	if !strings.HasPrefix(channel, "#") {
		channel = "#" + channel
	}

	return c.send("JOIN " + channel)
}

func (c *client) On(mt MessageType, f func(Message)) {
	c.handlers[mt] = f
}

// Part leaves channel
func (c *client) Part(channel string) error {
	if !strings.HasPrefix(channel, "#") {
		channel = "#" + channel
	}

	return c.send("PART " + channel)
}

// Say sends a PRIVMSG message in channel
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

func (c *client) handleMessage(rawMessage string) {
	ircData := parseIRCMessage(rawMessage)

	switch ircData.Prefix {
	case "tmi.twitch.tv":
		if h, ok := tmiTwitchTvHandlers(ircData.Command); ok {
			if h != nil {
				h(c, ircData)
			}
		} else {
			c.err = errors.New("could not handle message with tmi.twitch.tv prefix:\n" + rawMessage)
		}
	case "jtv":
		if h, ok := jtvHandlers(ircData.Command); ok {
			if h != nil {
				h(c, ircData)
			}
		} else {
			c.err = errors.New("could not handle message with jtv prefix:\n" + rawMessage)
		}
	default:
		if h, ok := otherHandlers(ircData.Command); ok {
			if h != nil {
				h(c, ircData)
			}
		} else {
			c.err = errors.New("could not handle message with { " + ircData.Prefix + " } as prefix:\n" + rawMessage)
		}
	}
}

func (c *client) readMessages() {
	for {
		c.rMutex.Lock()
		_, received, err := c.conn.ReadMessage()
		c.rMutex.Unlock()
		if err != nil {
			return
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
