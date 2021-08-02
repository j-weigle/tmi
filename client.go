package tmi

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type client struct {
	conn     *websocket.Conn
	config   *clientConfig
	done     chan bool
	err      chan error
	messages chan *Message
	mutex    *sync.Mutex
}

type clientConfig struct {
	Connection *connectionConfig // how the client will connect and reconnect
	Identity   *identityConfig   // who the client logs in as
	Channels   []string          // which channels the client will connect to
}

type connectionConfig struct {
	reconnect            bool // attempt to reconnect when disconnected?
	secure               bool // connect to to Twitch's secure server(port 443) or not (port 80)?
	reconnectInterval    int  // initial interval between reconnect attempts
	maxReconnectAttempts int  // maximum number of attempts to reconnect when disconnected
	maxReconnectInterval int  // maximum interval between reconnect attempts
}

type identityConfig struct {
	username string // login account name
	password string // oauth token
}

// NewClient returns a new client using config
func NewClient(config *clientConfig) *client {
	mutex := &sync.Mutex{}
	return &client{config: config, mutex: mutex}
}

// NewClientConfig returns a client config with Connection settings initialzed
// to the recommended defaults. Identity is initialzed but left empty.
func NewClientConfig() *clientConfig {
	var conn = &connectionConfig{}
	conn.SetToDefault()
	var id = &identityConfig{}
	return &clientConfig{Connection: conn, Identity: id}
}

// SetReconnect sets whether the client will attempt to reconnect
// to the server in the case of a disconnect.
func (c *connectionConfig) SetReconnect(reconnect bool) {
	c.reconnect = reconnect
}

// SetReconnectSettings sets how often and how many times the client
// will attempt to reconnect to the server in the case of a disconnect.
func (c *connectionConfig) SetReconnectSettings(interval, maxAttempts, maxInterval int) {
	c.reconnectInterval = interval
	c.maxReconnectAttempts = maxAttempts
	c.maxReconnectInterval = maxInterval
}

// SetSecure sets the connection scheme and port.
// true uses scheme = wss and port = 443.
// false uses scheme = ws and port = 80.
func (c *connectionConfig) SetSecure(secure bool) {
	c.secure = secure
}

// SetToDefault sets the connection configuration options to their recommended defaults.
//
// Default options:
// reconnect            = true,
// secure               = true,
// reconnectInterval    = 1000,
// maxReconnectAttempts = -1 (infinite),
// maxReconnectInterval = 30000 milliseconds,
func (c *connectionConfig) SetToDefault() {
	c.reconnect = true
	c.secure = true
	c.reconnectInterval = 1000
	c.maxReconnectAttempts = -1
	c.maxReconnectInterval = 30000
}

// Set sets the login identity configuration to username and password.
func (c *identityConfig) Set(username, password string) {
	if !strings.HasPrefix(password, "oauth:") {
		password = "oauth:" + password
	}
	c.username = username
	c.password = password
}

// SetToAnonymous sets username to an random justinfan username (password can be anything).
func (id *identityConfig) SetToAnonymous() {
	id.username = "justinfan" + fmt.Sprint(rand.Intn(79000)+1000)
	id.password = "swordfish"
}
