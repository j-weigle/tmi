package tmi

import (
	"errors"
	"math"
	"net/url"
	"strings"
	"time"
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
		case errFatalNotification:
			c.callDone(err)
		default:
			return err
		}
	}
}

// Disconnect closes the connection to the server, and does not attempt to reconnect
func (c *client) Disconnect() {
	c.notifDisconnect.notify() // let all relevant goroutines know the user called Disconnect
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

// OnDone sets the callback function to be called when a client is done (typically due to a fatal error).
func (c *client) OnDone(cb func(fatal error)) {
	c.done = cb
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
