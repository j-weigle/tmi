package tmi

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// see bottom of page at https://dev.twitch.tv/docs/irc/guide
const (
	// CapTags is for requesting the tags capability
	CapTags = "twitch.tv/tags"
	// CapCommands is for requesting the commands capability
	CapCommands = "twitch.tv/commands"
	// CapMembership is for requesting the membership capability
	CapMembership = "twitch.tv/membership"
)

// ClientConfig holds how a client connects, reconnects, logs in as, and pinger behavior.
type ClientConfig struct {
	Connection      ConnectionConfig // how the client will connect and reconnect
	Identity        IdentityConfig   // who the client logs in as
	Pinger          PingConfig       // how often to ping, and timeout
	Capabilities    []string         // which capabilites to request upon connection
	ReadBufferSize  int              // channel buffer size for inbound messages
	WriteBufferSize int              // channel buffer size for outbound messages
}

// ConnectionConfig holds reconnect settings and (in)secure server connection.
type ConnectionConfig struct {
	Reconnect            bool          // if true, reconnect on reconnect requests and non-fatal errors
	Secure               bool          // if true, connect to to Twitch's secure server(port 443), otherwise insecure (port 80)
	MaxReconnectAttempts int           // maximum number of attempts to reconnect when disconnected, -1 is infinite
	MaxReconnectInterval time.Duration // maximum interval between reconnect attempts
}

// IdentityConfig holds the username and password to log in with.
type IdentityConfig struct {
	Username string // login account name
	Password string // oauth token
}

// PingConfig holds whether or not to run a pinger, interval of the pinger, and timeout waiting on a pong.
type PingConfig struct {
	Enabled  bool          // whether to send pings or not
	Interval time.Duration // how long to wait before sending a ping when no messages have been received
	Timeout  time.Duration // how long to wait on a pong before reconnecting
}

// NewClientConfig returns a client config with Connection settings initialzed to the
// recommended defaults. Identity is set to username and password if not empty strings.
func NewClientConfig(username, password string) ClientConfig {
	conn := ConnectionConfig{}
	conn.Default()
	pinger := PingConfig{}
	pinger.Default()
	id := IdentityConfig{}
	if username != "" {
		id.SetUsername(username)
	}
	if password != "" {
		id.SetPassword(password)
	}
	return ClientConfig{
		Connection:      conn,
		Identity:        id,
		Pinger:          pinger,
		Capabilities:    []string{CapTags, CapCommands, CapMembership},
		ReadBufferSize:  512,
		WriteBufferSize: 512,
	}
}

// Default sets the connection configuration options to their recommended defaults.
// Default options:
// Reconnect            = true,
// Secure               = true,
// MaxReconnectAttempts = -1 (infinite),
// MaxReconnectInterval = 30 seconds,
func (c *ConnectionConfig) Default() {
	c.Reconnect = true
	c.Secure = true
	c.MaxReconnectAttempts = -1
	c.MaxReconnectInterval = time.Second * 30
}

// SetReconnectSettings sets how often and how many times the client
// will attempt to reconnect to the server in the case of a disconnect.
func (c *ConnectionConfig) SetReconnectSettings(maxAttempts int, maxInterval time.Duration) {
	c.MaxReconnectAttempts = maxAttempts
	if maxInterval < time.Second*5 {
		maxInterval = time.Second * 5
	}
	c.MaxReconnectInterval = maxInterval
}

// Anonymous sets username to an random justinfan username (password can be anything).
func (id *IdentityConfig) Anonymous() {
	id.Username = "justinfan" + fmt.Sprint(rand.Intn(79000)+1000)
	id.Password = "swordfish"
}

// Set sets the login identity configuration to username and password with oauth: prepended.
func (id *IdentityConfig) Set(username, password string) {
	id.SetUsername(username)
	id.SetPassword(password)
}

// SetPassword sets the password for the identity configuration to password with oauth: prepended.
func (id *IdentityConfig) SetPassword(password string) {
	password = strings.TrimSpace(password)
	if !strings.HasPrefix(password, "oauth:") {
		password = "oauth:" + password
	}
	id.Password = password
}

// SetUsername sets the username for the identity configuration to username.
func (id *IdentityConfig) SetUsername(username string) {
	username = strings.TrimSpace(username)
	id.Username = strings.ToLower(username)
}

// Default sets the ping configuration options to their recommended defaults.
func (p *PingConfig) Default() {
	p.Enabled = true
	p.Interval = time.Minute
	p.Timeout = time.Second * 5
}

// SetTimes sets the idle wait time and timeout for ping configuration.
func (p *PingConfig) SetTimes(interval, timeout time.Duration) {
	p.Interval = interval
	p.Timeout = timeout
}
