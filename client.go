package tmi

import (
	"fmt"
	"math/rand"
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

	// Done sets the callback function for when a client is done to cb (intended for use with fatal errors).
	Done(cb func(fatal error))

	// Join joins channel.
	Join(channel string) error

	// On sets the callback function to cb for the MessageType mt.
	On(mt MessageType, cb func(Message))

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
	conn             *websocket.Conn
	config           *clientConfig
	done             func(error)                   // callback function for fatal errors.
	handlers         map[MessageType]func(Message) // callback functions for each MessageType.
	inbound          chan string                   // for sending inbound messages to the handlers, mostly used for a buffer
	notifFatal       notifier                      // notification of fatal error.
	notifDisconnect  notifier                      // notification of user manual disconnect.
	notifReconnect   notifier                      // notification of reconnect.
	notifPingerDone  notifier                      // notification of pinger returning.
	onError          func(error)                   // callback function for non-fatal errors.
	outbound         chan string                   // for sending outbound messages to the writer
	rcvdMsg          chan bool                     // when conn reads, notifies ping loop.
	rcvdPong         chan bool                     // when pong received, notifies ping loop.
	reconnectCounter int                           // manages keeping track of reconnect attempts before a successful attempt
}

type clientConfig struct {
	Channels   []string          // which channels the client will connect to.
	Connection *connectionConfig // how the client will connect and reconnect.
	Identity   *identityConfig   // who the client logs in as.
	Pinger     *pingConfig       // how often to ping, and timeout.
}

type connectionConfig struct {
	reconnect            bool // if true, reconnect on reconnect requests and non-fatal errors.
	secure               bool // if true, connect to to Twitch's secure server(port 443), otherwise insecure (port 80).
	maxReconnectAttempts int  // maximum number of attempts to reconnect when disconnected.
	maxReconnectInterval int  // maximum interval between reconnect attempts.
}

type identityConfig struct {
	username string // login account name
	password string // oauth token
}

type pingConfig struct {
	wait    time.Duration // how long to wait before sending a ping when no messages have been received
	timeout time.Duration // how long to wait on a pong before reconnecting
}

// notifiers Reset() and Notify() are used in combination to notify multiple goroutines.
type notifier struct {
	mutex sync.Mutex
	once  *sync.Once
	ch    chan struct{}
}

// Reset sets the notifier to be ready to be used.
func (n *notifier) Reset() {
	n.mutex.Lock()
	n.once = &sync.Once{}
	n.ch = make(chan struct{})
	n.mutex.Unlock()
}

// Notify uses the notifier and makes it unusable until reset.
func (n *notifier) Notify() {
	n.mutex.Lock()
	n.once.Do(func() {
		close(n.ch)
	})
	n.mutex.Unlock()
}

// NewClient returns a new client using the provided config.
func NewClient(c *clientConfig) Client {
	var config = c.duplicate()
	var handlers = make(map[MessageType]func(Message))
	return &client{
		config:   config,
		handlers: handlers,
		inbound:  make(chan string, 512),
		outbound: make(chan string, 512),
		rcvdMsg:  make(chan bool),
	}
}

// duplicate returns a deep copy of the calling client config.
func (c *clientConfig) duplicate() *clientConfig {
	var config = &clientConfig{}
	var conn = *c.Connection
	var id = *c.Identity
	var chans = make([]string, len(c.Channels))
	var pinger = *c.Pinger
	copy(chans, c.Channels)
	config.Connection = &conn
	config.Identity = &id
	config.Channels = chans
	config.Pinger = &pinger
	return config
}

// NewClientConfig returns a client config with Connection settings initialzed
// to the recommended defaults. Identity is initialzed but left empty.
func NewClientConfig() *clientConfig {
	var conn = &connectionConfig{}
	conn.Default()
	var id = &identityConfig{}
	var pinger = &pingConfig{}
	pinger.Default()
	return &clientConfig{
		Connection: conn,
		Identity:   id,
		Pinger:     pinger,
	}
}

// SetReconnect sets whether the client will attempt to reconnect
// to the server in the case of a disconnect.
func (c *connectionConfig) SetReconnect(reconnect bool) {
	c.reconnect = reconnect
}

// SetReconnectSettings sets how often and how many times the client
// will attempt to reconnect to the server in the case of a disconnect.
func (c *connectionConfig) SetReconnectSettings(maxAttempts, maxInterval int) {
	c.maxReconnectAttempts = maxAttempts
	if maxInterval < 5000 {
		maxInterval = 5000
	}
	c.maxReconnectInterval = maxInterval
}

// SetSecure sets the connection scheme and port.
// true uses scheme = wss and port = 443.
// false uses scheme = ws and port = 80.
func (c *connectionConfig) SetSecure(secure bool) {
	c.secure = secure
}

// Default sets the connection configuration options to their recommended defaults.
//
// Default options:
// reconnect            = true,
// secure               = true,
// maxReconnectAttempts = -1 (infinite),
// maxReconnectInterval = 30000 milliseconds,
func (c *connectionConfig) Default() {
	c.reconnect = true
	c.secure = true
	c.maxReconnectAttempts = -1
	c.maxReconnectInterval = 30000
}

// Set sets the login identity configuration to username and password with oauth: prepended.
func (id *identityConfig) Set(username, password string) {
	id.SetUsername(username)
	id.SetPassword(password)
}

// SetPassword sets the password for the identity configuration to password with oauth: prepended.
func (id *identityConfig) SetPassword(password string) {
	if !strings.HasPrefix(password, "oauth:") {
		password = "oauth:" + password
	}
	id.password = password
}

// Anonymous sets username to an random justinfan username (password can be anything).
func (id *identityConfig) Anonymous() {
	id.username = "justinfan" + fmt.Sprint(rand.Intn(79000)+1000)
	id.password = "swordfish"
}

// SetUsername sets the username for the identity configuration to username.
func (id *identityConfig) SetUsername(username string) {
	id.username = strings.ToLower(username)
}

// Set sets the idle wait time and timeout for ping configuration p.
func (p *pingConfig) Set(wait, timeout time.Duration) {
	p.wait = wait
	p.timeout = timeout
}

// Default sets the ping configuration options to their recommended defaults.
func (p *pingConfig) Default() {
	p.wait = time.Minute
	p.timeout = time.Second * 5
}
