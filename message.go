package tmi

import (
	"time"
)

type MessageType int

const (
	// Unknown, unrecognized, or non-handled message types
	UNSET MessageType = iota - 1
	// tmi.twitch.tv prefixed message types
	CLEARCHAT
	CLEARMSG
	GLOBALUSERSTATE
	HOSTTARGET
	NOTICE
	RECONNECT
	ROOMSTATE
	USERNOTICE
	USERSTATE
	// non tmi.twitch.tv prefixed message types
	NAMES
	JOIN
	PART
	PING
	PONG
	PRIVMSG
	WHISPER
)

func (mt MessageType) String() string {
	if mt == -1 {
		return "UNSET"
	}
	return []string{
		"CLEARCHAT",
		"CLEARMSG",
		"GLOBALUSERSTATE",
		"HOSTTARGET",
		"NOTICE",
		"RECONNECT",
		"ROOMSTATE",
		"USERNOTICE",
		"USERSTATE",
		"NAMES",
		"JOIN",
		"PART",
		"PING",
		"PONG",
		"PRIVMSG",
		"WHISPER",
	}[mt]
}

type IRCTags map[string]string

type IRCData struct {
	Raw     string   `json:"raw"`
	Tags    IRCTags  `json:"tags"`
	Prefix  string   `json:"prefix"`
	Command string   `json:"command"`
	Params  []string `json:"params"`
}

type UnsetMessage struct {
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
}

// Timeout, ban, or clear all chat
type ClearChatMessage struct {
	Channel string      `json:"channel"`
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"` // a sentence explaining what the clear chat did
	Type    MessageType `json:"type"`

	BanDuration time.Duration `json:"ban-duration"` // duration of ban, omitted if permanent
	Target      string        `json:"target"`       // target of the ban, omitted if not a timeout or ban
}

// Singular message deletion
type ClearMsgMessage struct {
	Channel string      `json:"channel"`
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"` // the deleted message
	Type    MessageType `json:"type"`

	Login       string `json:"login"`         // name of user who sent deleted message
	TargetMsgID string `json:"target-msg-id"` // msg id of the deleted message
}

// Information about user that successfully logged in
type GlobalUserstateMessage struct {
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`

	EmoteSets []string `json:"emote-sets"` // emotes belonging to one or more emote sets
	User      *User    `json:"user"`       // information about the user that logged in
}

type HostTargetMessage struct {
	Channel string      `json:"channel"`
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"` // "<channel> is now hosting <hosted>" or "<channel> exited host mode"
	Type    MessageType `json:"type"`

	Hosted  string `json:"hosted"`  // the channel that was hosted by channel
	Viewers int    `json:"viewers"` // the number of viewers at channel during the host event
}

type NoticeMessage struct {
	Channel string      `json:"channel"`
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`

	Enabled bool     `json:"enabled"` // set when Notice is one of: emoteonly, uniquechat, subonly
	Mods    []string `json:"mods"`    // list of mods for Channel when Notice is set to mods
	MsgID   string   `json:"msg-id"`  // msg-id value from Data.Tags, or parse-error / login-error if the key doesn't exist
	Notice  string   `json:"notice"`  // Notice is one of: automod, emoteonly, mods, uniquechat, subonly, vips, notice (notice is the default)
	VIPs    []string `json:"vips"`    // list of vips for Channel when Notice is set to vips
}

type ReconnectMessage struct {
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`
}

type RoomstateMessage struct {
	Channel string      `json:"channel"`
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`

	// emote-only, followers-only, r9k(uniquechat), slow(slowmode), subs-only
	States map[string]RoomState `json:"states"` // the states in the roomstate tags
}

type RoomState struct { // note followers-only: -1 (disabled), 0 (enabled immediate chat), > 0 (enabled number of minutes delay to chat)
	Enabled bool          `json:"enabled"` // mode turned on or off
	Delay   time.Duration `json:"delay"`   // seconds between messages (slow), minutes post-follow (followers-only)
}

type UsernoticeMessage struct {
	Channel string      `json:"channel"`
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`

	Emotes    []Emote `json:"emotes"`     // parsed emotes string
	MsgParams IRCTags `json:"msg-params"` // any msg-param tags for the notice
	SystemMsg string  `json:"system-msg"` // message printed in chat on the notice
	User      *User   `json:"user"`       // user who caused the notice
}

type UserstateMessage struct {
	Channel string      `json:"channel"`
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`

	EmoteSets []string `json:"emote-sets"` // emotes belonging to one or more emote sets
	User      *User    `json:"user"`       // user that joined or sent a privmsg
}

type NamesMessage struct { // WARNING: deprecated, but not removed yet
	Channel string      `json:"channel"`
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`

	Users []string `json:"users"` // list of usernames
}

type JoinMessage struct {
	Channel string      `json:"channel"`
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`

	Username string `json:"username"` // name of joined account
}

type PartMessage struct {
	Channel string      `json:"channel"`
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`

	Username string `json:"username"` // name of parted account
}

type PingMessage struct {
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
}

type PongMessage struct {
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
}

type PrivateMessage struct {
	Channel string      `json:"channel"`
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`

	Action bool    `json:"action"` // indicates if the /me command was used
	Bits   int     `json:"bits"`   // number of bits if bits message
	Emotes []Emote `json:"emotes"` // parsed emotes string
	ID     string  `json:"id"`     // message id
	Reply  bool    `json:"reply"`  // indicates if the message is a reply
	User   *User   `json:"user"`   // user that sent the message
}

type ReplyMsgParent struct {
	DisplayName string `json:"display-name"`
	ID          string `json:"id"`   // message id
	Text        string `json:"text"` // message body
	UserID      string `json:"user-id"`
	Username    string `json:"username"` // login
}

type WhisperMessage struct {
	Channel string      `json:"channel"`
	Data    IRCData     `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`

	Action bool    `json:"action"`     // indicates if the /me command was used
	Emotes []Emote `json:"emotes"`     // parsed emotes string
	MsgID  string  `json:"message-id"` // tags["message-id"]
	Target string  `json:"target"`     // message recipient
	User   *User   `json:"user"`       // message sender
}

type Badge struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type Emote struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Positions []EmotePosition `json:"positions"`
}

type EmotePosition struct {
	EndIdx   int `json:"end-index"`
	StartIdx int `json:"start-index"`
}

type User struct {
	BadgeInfo   string  `json:"badge-info"`
	Badges      []Badge `json:"badges"`
	Broadcaster bool    `json:"broadcaster"`
	Color       string  `json:"color"`
	DisplayName string  `json:"display-name"`
	Mod         bool    `json:"mod"`
	Name        string  `json:"name"`
	Subscriber  bool    `json:"subscriber"`
	Turbo       bool    `json:"turbo"`
	ID          string  `json:"id"`
	UserType    string  `json:"user-type"`
	VIP         bool    `json:"vip"`
}
