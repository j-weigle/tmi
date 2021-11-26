# **tmi**
*Twitch Messaging Interface - Go Edition*

[![Gorilla Toolkit](https://www.gorillatoolkit.org/static/images/gorilla-icon-64.png)](https://www.gorillatoolkit.org) Powered by Gorilla Toolkit's [Gorilla WebSocket]

---

*Table of Contents*
- [**tmi**](#tmi)
	- [Overview](#overview)
	- [Features](#features)
	- [Getting Started](#getting-started)
	- [Messages](#messages)
		- [Message Types](#message-types)
		- [Message Data](#message-data)
	- [Client](#client)
		- [Client Methods](#client-methods)
		- [Client Event Callbacks](#client-event-callbacks)
	- [Configuration](#configuration)
		- [Configuration Options](#configuration-options)
		- [Configuration Methods](#configuration-methods)
	- [Rate Limiting](#rate-limiting)
		- [Adding a Join Rate Limiter](#adding-a-join-rate-limiter)
		- [Adding a Message Rate Limiter](#adding-a-message-rate-limiter)
		- [Presets](#presets)
		- [Rate Limit Methods and Types](#rate-limit-methods-and-types)
	- [Extra Parsing Functions/Methods](#extra-parsing-functionsmethods)

## Overview
TMI is a framework for getting a Go Twitch IRC bot up and running in no time using a websocket connection. It handles all the parsing and connection management for you, so you can stay focused on what you want your bot to do and forget about writing an IRC tag parser.

## Features
 - Simple, thread-safe API - run it blocking or non-blocking
 - Configurable rate limiting
 - Exponential reconnect backoff
 - Server pinging during inactivity
 - Tested common Twitch commands - skip writing your own timeout, ban, etc. functions
 - Special message variable parsing functions for tmi-sent-ts, reply, and system-msg

## Getting Started
```go
package main

import (
    "fmt"

    "github.com/j-weigle/tmi"
)

func main() {
    config := tmi.NewClientConfig("yourtwitchbotusername", "oauthtoken")
    client := tmi.NewClient(config)
    
    client.Join("channel")

    client.OnConnected(func() {
        client.Say("channel", "Hello, Chat!")
    })

    client.OnPrivateMessage(func(msg tmi.PrivateMessage) {
        fmt.Printf("%s says: '%s' in %s\n", msg.User.Name, msg.Text, msg.Channel)
    })
    
    err := client.Connect()
    if err != nil {
        panic(err)
    }
}
```
or alternatively for a non-blocking connection
```go
func main() {
    ...
    client.OnDone(func(err error) {
        panic(err)
    })
    
    go client.Connect()
    ...
}
```

## Messages

### Message Types
```go
UNSET
CLEARCHAT
CLEARMSG
GLOBALUSERSTATE
HOSTTARGET
NOTICE
RECONNECT
ROOMSTATE
USERNOTICE
USERSTATE
NAMES
JOIN
PART
PING
PONG
PRIVMSG
WHISPER
```

### Message Data
```go
// IRCTags for storing tags (when IRC message starts with @)
type IRCTags map[string]string

// IRCData for data parsed from raw IRC messages.
type IRCData struct {
	Raw     string
	Tags    IRCTags
	Prefix  string
	Command string
	Params  []string
}

// UnsetMessage for unrecognized or unset IRC commands or parsing errors.
type UnsetMessage struct {
	Data    IRCData
	IRCType string
	Text    string
	Type    MessageType
}

// ClearChatMessage when a timeout, ban, or clear all chat occurs.
type ClearChatMessage struct {
	Channel string
	Data    IRCData
	IRCType string
	Text    string
	Type    MessageType

	BanDuration time.Duration
	Target      string
}

// ClearMsgMessage data received on singular message deletion.
type ClearMsgMessage struct {
	Channel string
	Data    IRCData
	IRCType string
	Text    string
	Type    MessageType

	Login       string
	TargetMsgID string
}

// GlobalUserstateMessage data about user that successfully logged in.
type GlobalUserstateMessage struct {
	Data    IRCData
	IRCType string
	Type    MessageType

	EmoteSets []string
	User      *User
}

// HostTargetMessage data when a joined channel begins hosting another channel or exits host mode.
type HostTargetMessage struct {
	Channel string
	Data    IRCData
	IRCType string
	Text    string
	Type    MessageType

	Hosted  string
	Viewers int
}

// NoticeMessage data when a chat setting is changed, mods are received, login failures, etc.
type NoticeMessage struct {
	Channel string
	Data    IRCData
	IRCType string
	Text    string
	Type    MessageType

	Enabled bool
	Mods    []string
	MsgID   string
	Notice  string
	VIPs    []string
}

// ReconnectMessage data when the server requests that clients reconnect.
type ReconnectMessage struct {
	Data    IRCData
	IRCType string
	Type    MessageType
}

// RoomstateMessage data when a chat setting is changed, and includes the delay for certain settings.
type RoomstateMessage struct {
	Channel string
	Data    IRCData
	IRCType string
	Type    MessageType

	States map[string]RoomState
}

// RoomState data for a single room state
type RoomState struct {
	Enabled bool
	Delay   time.Duration
}

// UsernoticeMessage data when a user subscribes to a channel, incoming raid, and channel rituals.
type UsernoticeMessage struct {
	Channel string
	Data    IRCData
	IRCType string
	Text    string
	Type    MessageType

	Emotes    []Emote
	ID        string
	MsgID     string
	MsgParams IRCTags
	SystemMsg string
	User      *User
}

// UserstateMessage data when a user joins a channel or sends a PrivateMessage.
type UserstateMessage struct {
	Channel string
	Data    IRCData
	IRCType string
	Type    MessageType

	EmoteSets []string
	User      *User
}

// NamesMessage data when joining a channel, provides list of users in the chat.
type NamesMessage struct { // WARNING: deprecated, but not removed yet
	Channel string
	Data    IRCData
	IRCType string
	Type    MessageType

	Users []string
}

// JoinMessage data when user joins a channel, gives the channel joined and username joined as.
type JoinMessage struct {
	Channel string
	Data    IRCData
	IRCType string
	Type    MessageType

	Username string
}

// PartMessage data when user leaves a channel, gives the channel left and username that left.
type PartMessage struct {
	Channel string
	Data    IRCData
	IRCType string
	Type    MessageType

	Username string
}

// PingMessage data when a ping is received from the server.
type PingMessage struct {
	Data    IRCData
	IRCType string
	Text    string
	Type    MessageType
}

// PongMessage data when a pong is received from the server.
type PongMessage struct {
	Data    IRCData
	IRCType string
	Text    string
	Type    MessageType
}

// PrivateMessage data when a message is sent in a chat that is joined.
type PrivateMessage struct {
	Channel string
	Data    IRCData
	IRCType string
	Text    string
	Type    MessageType

	Action bool
	Bits   int
	Emotes []Emote
	ID     string
	Reply  bool
	User   *User
}

// ReplyParentMsg is the information provided in tags when a PrivateMessage is a reply.
type ReplyParentMsg struct {
	DisplayName string
	ID          string
	Text        string
	UserID      string
	Username    string
}

// WhisperMessage data when a whisper message is received.
type WhisperMessage struct {
	Data    IRCData
	IRCType string
	Text    string
	Type    MessageType

	Emotes []Emote
	ID     string
	Target string
	User   *User
}

// Badge represents a user chat badge badge/1
type Badge struct {
	Name  string
	Value int
}

// Emote information provided in tags.
type Emote struct {
	ID        string
	Name      string
	Positions []EmotePosition
}

// EmotePosition of emotes when emotes are in chat messages or whispers.
type EmotePosition struct {
	StartIdx int
	EndIdx   int
}

// User info provided in tags.
type User struct {
	BadgeInfo   string
	Badges      []Badge
	Broadcaster bool
	Color       string
	DisplayName string
	Mod         bool
	Name        string
	Subscriber  bool
	Turbo       bool
	ID          string
	UserType    string
	VIP         bool
}
```

## Client

### Client Methods
```go
func NewClient(c ClientConfig) *Client

func (c *Client) Connect() error
func (c *Client) Disconnect()
func (c *Client) Join(channels ...string) error
func (c *Client) Part(channels ...string) error
func (c *Client) Say(channel string, message string)

func (c *Client) Action(channel, message string) error
func (c *Client) Ban(channel, user, reason string) error
func (c *Client) Clear(channel string)
func (c *Client) Color(color string)
func (c *Client) Commercial(channel, seconds string)
func (c *Client) Delete(channel, messageID string)
func (c *Client) EmoteOnly(channel string)
func (c *Client) EmoteOnlyOff(channel string)
func (c *Client) Followers(channel, duration string)
func (c *Client) FollowersOff(channel string)
func (c *Client) Host(channel, target string)
func (c *Client) Marker(channel, description string) error
func (c *Client) Mod(channel, user string)
func (c *Client) Mods(channel string)
func (c *Client) R9kBeta(channel string)
func (c *Client) R9kBetaOff(channel string)
func (c *Client) R9kMode(channel string)
func (c *Client) R9kModeOff(channel string)
func (c *Client) Raid(channel, target string)
func (c *Client) Slow(channel, seconds string)
func (c *Client) SlowOff(channel string)
func (c *Client) Subscribers(channel string)
func (c *Client) SubscribersOff(channel string)
func (c *Client) Timeout(channel, user, seconds string)
func (c *Client) UnVIP(channel, user string)
func (c *Client) Unban(channel, user string)
func (c *Client) Unhost(channel string)
func (c *Client) Uniquechat(channel string)
func (c *Client) UniquechatOff(channel string)
func (c *Client) Unmod(channel, user string)
func (c *Client) Unraid(channel string)
func (c *Client) Untimeout(channel, user string)
func (c *Client) VIP(channel, user string)
func (c *Client) VIPs(channel string)

func (c *Client) SetJoinRateLimit(rl RateLimit)
func (c *Client) UpdatePassword(password string)
```

### Client Event Callbacks
```go
func (c *Client) OnDone(cb func(fatal error))
func (c *Client) OnUnsetMessage(cb func(UnsetMessage))
func (c *Client) OnConnected(cb func())
func (c *Client) OnClearChatMessage(cb func(ClearChatMessage))
func (c *Client) OnClearMsgMessage(cb func(ClearMsgMessage))
func (c *Client) OnGlobalUserstateMessage(cb func(GlobalUserstateMessage))
func (c *Client) OnHostTargetMessage(cb func(HostTargetMessage))
func (c *Client) OnNoticeMessage(cb func(NoticeMessage))
func (c *Client) OnReconnectMessage(cb func(ReconnectMessage))
func (c *Client) OnRoomstateMessage(cb func(RoomstateMessage))
func (c *Client) OnUserNoticeMessage(cb func(UsernoticeMessage))
func (c *Client) OnUserstateMessage(cb func(UserstateMessage))
func (c *Client) OnNamesMessage(cb func(NamesMessage))
func (c *Client) OnJoinMessage(cb func(JoinMessage))
func (c *Client) OnPartMessage(cb func(PartMessage))
func (c *Client) OnPingMessage(cb func(PingMessage))
func (c *Client) OnPongMessage(cb func(PongMessage))
func (c *Client) OnPrivateMessage(cb func(PrivateMessage))
func (c *Client) OnWhisperMessage(cb func(WhisperMessage))
```

## Configuration

### Configuration Options
*Note that it is recommended to change configuration options through the methods provided as they do things like make sure oauth: is prepended to your oauth token, but it is not necessary.*
```go
// ClientConfig.Capabilites options
const (
	CapTags = "twitch.tv/tags"
	CapCommands = "twitch.tv/commands"
	CapMembership = "twitch.tv/membership"
)

type ClientConfig struct {
	Connection      ConnectionConfig
	Identity        IdentityConfig
	Pinger          PingConfig
	Capabilities    []string
	ReadBufferSize  int
	WriteBufferSize int
}

type ConnectionConfig struct {
	Reconnect            bool
	Secure               bool
	MaxReconnectAttempts int // -1 for infinite
	MaxReconnectInterval time.Duration
}

type IdentityConfig struct {
	Username string
	Password string
}

type PingConfig struct {
	Enabled  bool
	Interval time.Duration
	Timeout  time.Duration
}
```

### Configuration Methods
```go
func NewClientConfig(username, password string) ClientConfig

func (c *ConnectionConfig) Default()
func (c *ConnectionConfig) SetReconnectSettings(maxAttempts int, maxInterval time.Duration)

func (id *IdentityConfig) Anonymous()
func (id *IdentityConfig) Set(username, password string)
func (id *IdentityConfig) SetPassword(password string)
func (id *IdentityConfig) SetUsername(username string)

func (p *PingConfig) Default()
func (p *PingConfig) SetTimes(interval, timeout time.Duration)
```

## Rate Limiting

### Adding a Join Rate Limiter
```go
func main() {
	...
	client.SetJoinRateLimit(tmi.RLimJoinDefault)
	...
}
```
or make a custom RateLimit
```go
// 5 attempts per 10 seconds
rl := tmi.RateLimit{
	Burst: 5,
	Rate: time.Second * 10/5
}
```

### Adding a Message Rate Limiter
*This is a very crude example, but shows how to get a rate limiter and use Wait(). Wait is thread safe, therefore you are able to use it for multiple goroutines.*
```go
func main() {
	...
	rlimiter := tmi.NewRateLimiter(tmi.RLimMsgDefault)
	client.OnPrivateMessage(func(msg tmi.PrivateMessage) {
		rlimiter.Wait()
		client.Say("channel", "Message Received")
	})
	...
}

```

### Presets

```go
// RLimJoinDefault is the regular account rate limit 20 attempts per 10s
RLimJoinDefault = RateLimit{Burst: 10, Rate: time.Second * 10 / 20}

// RLimJoinVerified is the verified account rate limit 2000 attempts per 10s
RLimJoinVerified = RateLimit{Burst: 1000, Rate: time.Second * 10 / 2000}

// RLimMsgDefault is the regular account rate limit 20 messages per 30s
RLimMsgDefault = RateLimit{Burst: 10, Rate: time.Second * 30 / 20}

// RLimMsgMod is the mod/broadcaster/VIP account rate limit 100 messages per 30s
RLimMsgMod = RateLimit{Burst: 50, Rate: time.Second * 30 / 100}

// RLimGlobalDefault is the verified account rate limit 7500 global messages per 30s
// assumed to be global limit for other accounts as well since it is undocumented
RLimGlobalDefault = RateLimit{Burst: 3750, Rate: time.Second * 30 / 7500}

// RLimWhisperDefault is the rate limit for any account of 100 messages per 60s
// 100 / minute is more constricting than 3 / second, so it is chosen
RLimWhisperDefault = RateLimit{Burst: 2, Rate: time.Minute / 100}
```

### Rate Limit Methods and Types
```go
func NewRateLimiter(rl RateLimit) *RateLimiter
func (rl *RateLimiter) Wait()

type RateLimit struct {
	Burst int
	Rate time.Duration
}
```

## Extra Parsing Functions/Methods
```go
// EscapeIRCTagValues escapes strings in certain messages' IRCTags `\s` -> " ", `\n` -> "\n", `\r` -> "\r", `\:` -> ";", `\\` -> "\\"
// for example, the system-msg tag of UserNotice messages sometimes has these symbols
func (tags IRCTags) EscapeIRCTagValues()

// Same as above, but a function rather than a method of IRCTags
func EscapeIRCTagValues(tags ...string) []string

// This is for tmi-sent-ts, etc
func ParseTimeStamp(unixTime string) time.Time

// This is for when a PrivateMessage has its Reply field set to true
func ParseReplyParentMessage(tags IRCTags) ReplyParentMsg
```

[gorilla websocket]: <https://github.com/gorilla/websocket>