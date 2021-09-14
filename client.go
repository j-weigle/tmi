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

const pingSignature = "go-tmi-ws"

// Client to configure callbacks and manage the connection.
type Client struct {
	channels         map[string]bool
	channelsMutex    sync.Mutex
	config           ClientConfig
	conn             *websocket.Conn
	connected        atomicBool
	done             func(error) // callback function for fatal errors.
	handlers         onMessageHandlers
	inbound          chan string   // for sending inbound messages to the handlers, acts as a buffer.
	notifDisconnect  notifier      // used for disconnect call notifications
	outbound         chan string   // for sending outbound messages to the writer.
	rcvdMsg          chan struct{} // when conn reads, notifies ping loop.
	rcvdPong         chan struct{} // when pong received, notifies ping loop.
	reconnectCounter int           // for keeping track of reconnect attempts before a successful attempt.
	rLimiterJoins    *RateLimiter
}

type onMessageHandlers struct {
	onUnsetMessage           func(UnsetMessage)
	onConnected              func()
	onClearChatMessage       func(ClearChatMessage)
	onClearMsgMessage        func(ClearMsgMessage)
	onGlobalUserstateMessage func(GlobalUserstateMessage)
	onHostTargetMessage      func(HostTargetMessage)
	onNoticeMessage          func(NoticeMessage)
	onReconnectMessage       func(ReconnectMessage)
	onRoomstateMessage       func(RoomstateMessage)
	onUserNoticeMessage      func(UsernoticeMessage)
	onUserstateMessage       func(UserstateMessage)
	onNamesMessage           func(NamesMessage)
	onJoinMessage            func(JoinMessage)
	onPartMessage            func(PartMessage)
	onPingMessage            func(PingMessage)
	onPongMessage            func(PongMessage)
	onPrivateMessage         func(PrivateMessage)
	onWhisperMessage         func(WhisperMessage)
}

// NewClient returns a new client using the provided config.
func NewClient(c ClientConfig) *Client {
	return &Client{
		channels: make(map[string]bool),
		config:   c,
		inbound:  make(chan string, 512),
		outbound: make(chan string, 512),
		rcvdMsg:  make(chan struct{}),
	}
}

func (c *Client) callDone(err error) {
	if c.done != nil {
		c.done(err)
	}
}

func (c *Client) connect(u url.URL) error {
	var err error
	select { // Check disconnect has not been called before bothering to reconnect.
	case <-c.notifDisconnect.ch:
		return ErrDisconnectCalled
	default:
	}

	// Establish a connection to the URL defined by u.
	if c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil); err != nil {
		return errReconnect
	}

	// Waitgroup and context for goroutine control.
	var wg = &sync.WaitGroup{}
	var ctx, cancelFunc = context.WithCancel(context.Background())

	var closeErr = &connCloseErr{}
	// Let goroutines have a callback to signal one another to return using context's CancelFunc.
	var closeErrCb = func(errReason error) {
		cancelFunc()
		c.connected.set(false)
		closeErr.update(errReason)
	}

	// Begin reading from c.conn in separate goroutine.
	c.spawnReader(ctx, wg, closeErrCb)

	// Send NICK, PASS, and CAP REQ.
	// Sends in this goroutine before starting writer to prevent write conflicts.
	err = c.sendConnectSequence()
	if err != nil {
		closeErrCb(errReconnect)
	}

	// Begin writing to c.conn in separate goroutine.
	c.spawnWriter(ctx, wg, closeErrCb)

	// Start the pinger in a separate goroutine.
	// It will ping c.conn after it hasn't received a message for c.config.Pinger.interval.
	if c.config.Pinger.enabled {
		c.spawnPinger(ctx, wg, closeErrCb)
	}

	// Block and wait for a disconnect call or a connection error.
	c.listenAndParse(ctx, closeErrCb)

	// Make sure reader, writer, and pinger have finished.
	wg.Wait()

	return closeErr.err
}

// disconnect sends a close message to the server and then closes the connection.
func (c *Client) disconnect() bool {
	defer c.conn.Close()
	return c.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")) == nil
}

// sends joins using rate limiter if one is set
func (c *Client) joinChannels(channels []string) {
	if channels == nil || len(channels) < 1 {
		return
	}

	for _, ch := range channels {
		if c.rLimiterJoins != nil {
			c.rLimiterJoins.Wait()
		}
		if !c.connected.get() {
			return
		}
		if c.send("JOIN "+ch) == nil {
			c.channelsMutex.Lock()
			c.channels[ch] = c.connected.get()
			c.channelsMutex.Unlock()
		}
	}
}

func (c *Client) onConnectedJoins() {
	var channels = []string{}
	c.channelsMutex.Lock()
	for channel := range c.channels {
		c.channels[channel] = false
		channels = append(channels, channel)
	}
	c.channelsMutex.Unlock()
	c.joinChannels(channels)
}

func (c *Client) listenAndParse(ctx context.Context, closeErrCb func(error)) {
	for {
		select {
		case <-c.notifDisconnect.ch:
			closeErrCb(ErrDisconnectCalled)
			return
		case <-ctx.Done():
			return
		case rawMessage := <-c.inbound:
			if err := c.handleIRCMessage(rawMessage); err != nil {
				closeErrCb(err)
				return
			}
		}
	}
}

func (c *Client) send(message string) error {
	select {
	case c.outbound <- message:
		return nil
	default:
		return errors.New("message not delivered to outbound channel")
	}
}

func (c *Client) sendConnectSequence() (err error) {
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

func (c *Client) spawnPinger(ctx context.Context, wg *sync.WaitGroup, closeErrCb func(error)) {
	// recreate each time so that there isn't a pong sitting in the channel on reconnects
	c.rcvdPong = make(chan struct{}, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			var intervalT = time.NewTimer(c.config.Pinger.interval)
			select {
			case <-ctx.Done():
				if !intervalT.Stop() {
					<-intervalT.C
				}
				return

			case <-c.rcvdMsg:
				if !intervalT.Stop() {
					<-intervalT.C
				}

			case <-intervalT.C:
				err := c.send("PING :" + pingSignature)
				if err != nil {
					closeErrCb(errReconnect)
				}

				var timeoutT = time.NewTimer(c.config.Pinger.timeout)
				select {
				case <-c.rcvdPong:
					if !timeoutT.Stop() {
						<-timeoutT.C
					}

				case <-timeoutT.C:
					closeErrCb(errReconnect)
					return
				}
			}
		}
	}()
}

func (c *Client) spawnReader(ctx context.Context, wg *sync.WaitGroup, closeErrCb func(error)) {
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

func (c *Client) spawnWriter(ctx context.Context, wg *sync.WaitGroup, closeErrCb func(error)) {
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
