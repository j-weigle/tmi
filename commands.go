package tmi

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

const (
	wssServ = "irc-ws.chat.twitch.tv:443"
	wsServ  = "irc-ws.chat.twitch.tv:80"
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
		u = url.URL{Scheme: "wss", Host: wssServ}
	} else {
		u = url.URL{Scheme: "ws", Host: wsServ}
	}

	c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	c.done = make(chan bool)
	c.err = make(chan error)
	c.messages = make(chan *Message, 50)

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
	c.mutex.Lock()
	err := c.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.mutex.Unlock()
	return err
}

// Join joins channel
func (c *client) Join(channel string) error {
	if !strings.HasPrefix(channel, "#") {
		channel = "#" + channel
	}

	return c.send("JOIN " + channel)
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
		return fmt.Errorf("cannot send messages as an anonymous user")
	}
	if len(message) >= 500 {
		return fmt.Errorf("twitch chat's message length limit is 500 characters")
	}

	if !strings.HasPrefix(channel, "#") {
		channel = "#" + channel
	}
	return c.send("PRIVMSG " + channel + " :" + message)
}

func (c *client) handleMessage(rawMessage string) {
	msgdata := parseMessage(rawMessage)

	if msgdata.Prefix == "tmi.twitch.tv" {
		if f, ok := tmiTwitchTvCommands[msgdata.Command]; ok && f != nil {
			f(c, msgdata)
		}
	} else if msgdata.Prefix == "jtv" {
		if f, ok := jtvCommands[msgdata.Command]; ok && f != nil {
			f(c, msgdata)
		}
	} else {
		if f, ok := otherCommands[msgdata.Command]; ok && f != nil {
			f(c, msgdata)
		}
	}
}

func (c *client) readMessages() {
	defer close(c.done)
	defer close(c.err)
	defer close(c.messages)

	for {
		_, received, err := c.conn.ReadMessage()
		if err != nil {
			c.err <- err
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

func (c *client) send(message string) error {
	c.mutex.Lock()
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(message))
	c.mutex.Unlock()
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
