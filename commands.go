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
	errReconnect = errors.New("reconnect")
	// ErrDisconnectCalled is returned from Connect and in OnDone when the client calls disconnect.
	ErrDisconnectCalled = errors.New("disconnect was called")
	// ErrLoginFailure is returned from Connect and in OnDone when the client receives a NOTICE message about a login failure.
	ErrLoginFailure = errors.New("login failure")
	// ErrMaxReconnectAttemptsReached is returned from Connect and in OnDone when the client has attempted to reconnect the maximum number of times alloted by its config.
	ErrMaxReconnectAttemptsReached = errors.New("max attempts to reconnect reached")
)

// Connect connects to irc-ws.chat.twitch.tv and attempts to reconnect on connection errors.
func (c *Client) Connect() error {
	var err error
	var u url.URL

	if c.config.Connection.secure {
		u = url.URL{Scheme: "wss", Host: twitchWSSHost}
	} else {
		u = url.URL{Scheme: "ws", Host: twitchWSHost}
	}

	var maxReconnectAttempts int = c.config.Connection.maxReconnectAttempts
	var maxReconnectInterval time.Duration = c.config.Connection.maxReconnectInterval

	// Reset disconnect before starting connection loop. connect() will check if it has
	// been used before attempting to (re)connect.
	c.notifDisconnect.reset()

	for {
		err = c.connect(u)
		if c.config.Connection.reconnect {
			switch err {
			case errReconnect:
				const overflowPoint = 64 // technically 63, but using i - 1

				var i int = c.reconnectCounter
				c.reconnectCounter++
				// in case of overflow, reset to overflow point in order to maintain max interval
				if c.reconnectCounter < 0 {
					c.reconnectCounter = overflowPoint
				}

				if maxReconnectAttempts >= 0 && i >= maxReconnectAttempts {
					c.callDone(ErrMaxReconnectAttemptsReached)
					return err
				}

				var sleepDuration time.Duration
				if i == 0 {
					continue // immediate reconnect on first attempt
				} else if i > 0 && i < overflowPoint {
					// i - 1 because math.Pow(2, 0) == 1
					sleepDuration = time.Duration(math.Pow(2, float64(i-1)))
				} else {
					sleepDuration = maxReconnectInterval
				}

				if sleepDuration > maxReconnectInterval {
					sleepDuration = maxReconnectInterval
				}

				time.Sleep(sleepDuration)

			default:
				c.callDone(err)
				return err
			}
		}
	}
}

// Disconnect closes the connection to the server, and does not attempt to reconnect.
func (c *Client) Disconnect() {
	c.notifDisconnect.notify()
}

// Join joins channels.
func (c *Client) Join(channels ...string) error {
	if channels == nil || len(channels) < 1 {
		return errors.New("channels was empty or nil")
	}

	var newJoins = []string{}
	c.channelsMutex.Lock()
	for _, channel := range channels {
		channel = formatChannel(channel)

		connected, ok := c.channels[channel]
		if !ok {
			c.channels[channel] = false
		}
		if !connected {
			newJoins = append(newJoins, channel)
		}
	}
	c.channelsMutex.Unlock()

	if c.connected.get() {
		if len(newJoins) > 0 {
			go c.joinChannels(newJoins)
		}
	}
	return nil
}

// Part leaves channels.
func (c *Client) Part(channels ...string) error {
	if channels == nil || len(channels) < 1 {
		return errors.New("channels was empty or nil")
	}

	for _, channel := range channels {
		channel = formatChannel(channel)

		c.channelsMutex.Lock()
		delete(c.channels, channel)
		c.channelsMutex.Unlock()

		if c.connected.get() {
			c.send("PART " + channel)
		}
	}

	return nil
}

// Say sends a PRIVMSG message in channel.
func (c *Client) Say(channel string, message string) {
	channel = formatChannel(channel)

	if len(message) < 500 {
		c.send("PRIVMSG " + channel + " :" + message)
		return
	}
	var messages = splitChatMessage(message)
	for _, m := range messages {
		c.send("PRIVMSG " + channel + " :" + m)
	}
}

// Action sends a message as a /me, or action, message.
func (c *Client) Action(channel, message string) error {
	if len(message) > 490 {
		return errors.New("message must be shorter than 490 characters")
	}
	c.Say(channel, "\u0001ACTION "+message+"\u0001")
	return nil
}

// Ban bans user from reading or sending messages in channel with optional reason.
func (c *Client) Ban(channel, user, reason string) error {
	if len(reason)+len(user) > 490 {
		return errors.New("user + reason must be shorter than 490 characters")
	}
	if reason != "" {
		c.Say(channel, "/ban "+user+" "+reason)
		return nil
	}
	c.Say(channel, "/ban "+user)
	return nil
}

// Unban unbans user from channel.
func (c *Client) Unban(channel, user string) {
	c.Say(channel, "/unban "+user)
}

// Clear clears all chat messages in channel.
func (c *Client) Clear(channel string) {
	c.Say(channel, "/clear")
}

// Color changes the color of the username currently logged in.
func (c *Client) Color(color string) {
	c.Say("#"+c.config.Identity.username, "/color "+color)
}

// Commercial starts a a commercial break in channel that is seconds long.
// seconds should be 30, 60, 90, 120, 150, or 180.
func (c *Client) Commercial(channel, seconds string) {
	c.Say(channel, "/commercial "+seconds)
}

// Delete deletes a single message in channel identified by messageID.
// messageID for a PrivateMessage is PrivateMessage.ID.
// messageID for a ReplyParentMsg is ReplyParentMsg.ID.
func (c *Client) Delete(channel, messageID string) {
	c.Say(channel, "/delete "+messageID)
}

// EmoteOnly turns on emoteonly mode in channel.
func (c *Client) EmoteOnly(channel string) {
	c.Say(channel, "/emoteonly")
}

// EmoteOnlyOff turns off emoteonly mode in channel.
func (c *Client) EmoteOnlyOff(channel string) {
	c.Say(channel, "/emoteonlyoff")
}

// Followers turns on followersonly mode in channel with duration being how long a
// user must be following before they can send messages.
func (c *Client) Followers(channel, duration string) {
	c.Say(channel, "/followers "+duration)
}

// FollowersOff turns off followersonly mode in channel.
func (c *Client) FollowersOff(channel string) {
	c.Say(channel, "/followersoff")
}

// Host starts hosting target in channel. Trims off # from beginning of target.
func (c *Client) Host(channel, target string) {
	target = strings.TrimPrefix(target, "#")
	c.Say(channel, "/host "+target)
}

// Unhost stops hosting in channel.
func (c *Client) Unhost(channel string) {
	c.Say(channel, "/unhost")
}

// Marker adds a stream marker in channel with optional description.
func (c *Client) Marker(channel, description string) error {
	if len(description) > 490 {
		return errors.New("description must be shorter than 490 characters")
	}
	if description != "" {
		c.Say(channel, "/marker "+description)
		return nil
	}
	c.Say(channel, "/marker")
	return nil
}

// Mod makes user a moderator in channel.
func (c *Client) Mod(channel, user string) {
	c.Say(channel, "/mod "+user)
}

// Unmod makes user no longer a moderator in channel.
func (c *Client) Unmod(channel, user string) {
	c.Say(channel, "/unmod "+user)
}

// Mods requests the list of mods for channel. Use OnNoticeMessage to get the result.
// NoticeMessage.Notice of type "mods" indicates that NoticeMessage.Mods contains the result.
func (c *Client) Mods(channel string) {
	c.Say(channel, "/mods")
}

// R9kBeta turns on r9kbeta(uniquechat) mode in channel.
func (c *Client) R9kBeta(channel string) {
	c.Say(channel, "/r9kbeta")
}

// R9kMode turns on r9kbeta(uniquechat) mode in channel.
func (c *Client) R9kMode(channel string) {
	c.R9kBeta(channel)
}

// Uniquechat turns on uniquechat(r9kbeta) mode in channel.
func (c *Client) Uniquechat(channel string) {
	c.R9kBeta(channel)
}

// R9kBetaOff turns off r9kbeta(uniquechat) mode in channel.
func (c *Client) R9kBetaOff(channel string) {
	c.Say(channel, "/r9kbetaoff")
}

// R9kModeOff turns off r9kbeta(uniquechat) mode in channel.
func (c *Client) R9kModeOff(channel string) {
	c.R9kBetaOff(channel)
}

// UniquechatOff turns off uniquechat(r9kbeta) mode in channel.
func (c *Client) UniquechatOff(channel string) {
	c.R9kBetaOff(channel)
}

// Raid starts a raid on channel to target. Trims off # from beginning of target.
func (c *Client) Raid(channel, target string) {
	target = strings.TrimPrefix(target, "#")
	c.Say(channel, "/raid "+target)
}

// Unraid cancels a raid on channel.
func (c *Client) Unraid(channel string) {
	c.Say(channel, "/unraid")
}

// Slow turns on slow mode in channel with seconds delay between users sending messages.
func (c *Client) Slow(channel, seconds string) {
	c.Say(channel, "/slow "+seconds)
}

// SlowOff turns off slow mode in channel.
func (c *Client) SlowOff(channel string) {
	c.Say(channel, "/slowoff")
}

// Subscribers turns on subscribers only mode in channel.
func (c *Client) Subscribers(channel string) {
	c.Say(channel, "/subscribers")
}

// SubscribersOff turns off subscribers only mode in channel.
func (c *Client) SubscribersOff(channel string) {
	c.Say(channel, "/subscribersoff")
}

// Timeout prevents user in channel from chatting for seconds and clears their messsages.
func (c *Client) Timeout(channel, user, seconds string) {
	if seconds != "" {
		c.Say(channel, "/timeout "+user+" "+seconds)
		return
	}
	c.Say(channel, "/timeout "+user)
}

// Untimeout removes a timeout for user in channel.
func (c *Client) Untimeout(channel, user string) {
	c.Say(channel, "/untimeout "+user)
}

// VIP makes user a vip in channel.
func (c *Client) VIP(channel, user string) {
	c.Say(channel, "/vip "+user)
}

// UnVIP makes user no longer a vip in channel.
func (c *Client) UnVIP(channel, user string) {
	c.Say(channel, "/unvip "+user)
}

// VIPs requests the list of vips for channel. Use OnNoticeMessage to get the result.
// NoticeMessage.Notice of type "vips" indicates that NoticeMessage.VIPs contains the result.
func (c *Client) VIPs(channel string) {
	c.Say(channel, "/vips")
}

// SetJoinRateLimit sets the RateLimiter for JOIN commands to settings in RateLimit.
func (c *Client) SetJoinRateLimit(rl RateLimit) {
	c.rLimiterJoins = NewRateLimiter(rl)
}

// UpdatePassword updates the password the client uses for authentication.
func (c *Client) UpdatePassword(password string) {
	c.config.Identity.SetPassword(password)
}

// OnDone sets the callback function for when a client is done to cb. Useful for running a client in a goroutine.
func (c *Client) OnDone(cb func(fatal error)) {
	c.done = cb
}

// OnUnsetMessage sets the callback for when an unrecognized, non-handled, or unparsable message type is received.
func (c *Client) OnUnsetMessage(cb func(UnsetMessage)) {
	c.handlers.onUnsetMessage = cb
}

// OnConnected sets the callback for when the client successfully connects.
func (c *Client) OnConnected(cb func()) {
	c.handlers.onConnected = cb
}

// OnClearChatMessage sets the callback for when a CLEARCHAT message is received.
func (c *Client) OnClearChatMessage(cb func(ClearChatMessage)) {
	c.handlers.onClearChatMessage = cb
}

// OnClearMsgMessage sets the callback for when a CLEARMSG message is received.
func (c *Client) OnClearMsgMessage(cb func(ClearMsgMessage)) {
	c.handlers.onClearMsgMessage = cb
}

// OnGlobalUserstateMessage sets the callback for when a GLOBALUSERSTATE message is received.
func (c *Client) OnGlobalUserstateMessage(cb func(GlobalUserstateMessage)) {
	c.handlers.onGlobalUserstateMessage = cb
}

// OnHostTargetMessage sets the callback for when a HOSTTARGET message is received.
func (c *Client) OnHostTargetMessage(cb func(HostTargetMessage)) {
	c.handlers.onHostTargetMessage = cb
}

// OnNoticeMessage sets the callback for when a NOTICE message is received.
func (c *Client) OnNoticeMessage(cb func(NoticeMessage)) {
	c.handlers.onNoticeMessage = cb
}

// OnReconnectMessage sets the callback for when a RECONNECT message is received.
func (c *Client) OnReconnectMessage(cb func(ReconnectMessage)) {
	c.handlers.onReconnectMessage = cb
}

// OnRoomstateMessage sets the callback for when a ROOMSTATE message is received.
func (c *Client) OnRoomstateMessage(cb func(RoomstateMessage)) {
	c.handlers.onRoomstateMessage = cb
}

// OnUserNoticeMessage sets the callback for when a USERNOTICE message is received.
func (c *Client) OnUserNoticeMessage(cb func(UsernoticeMessage)) {
	c.handlers.onUserNoticeMessage = cb
}

// OnUserstateMessage sets the callback for when a USERSTATE message is received.
func (c *Client) OnUserstateMessage(cb func(UserstateMessage)) {
	c.handlers.onUserstateMessage = cb
}

// OnNamesMessage sets the callback for when a 353 message is received.
func (c *Client) OnNamesMessage(cb func(NamesMessage)) {
	c.handlers.onNamesMessage = cb
}

// OnJoinMessage sets the callback for when a JOIN message is received.
func (c *Client) OnJoinMessage(cb func(JoinMessage)) {
	c.handlers.onJoinMessage = cb
}

// OnPartMessage sets the callback for when a PART message is received.
func (c *Client) OnPartMessage(cb func(PartMessage)) {
	c.handlers.onPartMessage = cb
}

// OnPingMessage sets the callback for when a PING message is received.
func (c *Client) OnPingMessage(cb func(PingMessage)) {
	c.handlers.onPingMessage = cb
}

// OnPongMessage sets the callback for when a PONG message is received.
func (c *Client) OnPongMessage(cb func(PongMessage)) {
	c.handlers.onPongMessage = cb
}

// OnPrivateMessage sets the callback for when a PRIVMSG message is received.
func (c *Client) OnPrivateMessage(cb func(PrivateMessage)) {
	c.handlers.onPrivateMessage = cb
}

// OnWhisperMessage sets the callback for when a WHISPER message is received.
func (c *Client) OnWhisperMessage(cb func(WhisperMessage)) {
	c.handlers.onWhisperMessage = cb
}

func formatChannel(channel string) string {
	channel = strings.TrimSpace(channel)
	if !strings.HasPrefix(channel, "#") {
		channel = "#" + channel
	}
	return strings.ToLower(channel)
}

func splitChatMessage(message string) []string {
	const splIdx = 500
	var messages []string

	for len(message) >= splIdx {
		var lastSpace = strings.LastIndex(message[:splIdx], " ")
		if lastSpace == -1 {
			lastSpace = splIdx
		}
		messages = append(messages, strings.TrimSpace(message[:lastSpace]))
		message = strings.TrimSpace(message[lastSpace:])
	}
	if message != "" {
		messages = append(messages, message)
	}

	return messages
}
