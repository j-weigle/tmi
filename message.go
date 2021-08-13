package tmi

import "time"

type MessageType int

const (
	// Uknown, unrecognized, or non-handled message types
	UNSET MessageType = iota - 1
	// tmi.twitch.tv prefixed message types
	WELCOME
	INVALIDIRC
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
		"WELCOME",
		"INVALIDIRC",
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

type Message interface {
	GetType() MessageType
}

type IRCData struct {
	Raw     string            `json:"raw"`
	Tags    map[string]string `json:"tags"`
	Prefix  string            `json:"prefix"`
	Command string            `json:"command"`
	Params  []string          `json:"params"`
}

type UnsetMessage struct {
	Data    *IRCData
	IRCType string
	Text    string
	Type    MessageType
}

// Initial welcome message after successfully loggging in
type WelcomeMessage struct {
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`
}

type InvalidIRCMessage struct {
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"` // "Unknown command"
	Type    MessageType `json:"type"`

	Unknown string `json:"uknown"` // the command that was used
	User    string `json:"user"`   // the user who issued the command
}

// Timeout, ban, or clear all chat
type ClearChatMessage struct {
	Channel string      `json:"channel"`
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"` // a sentence explaining what the clear chat did
	Type    MessageType `json:"type"`

	BanDuration time.Duration `json:"ban-duration"` // duration of ban, omitted if permanent
	Target      string        `json:"target"`       // target of the ban, omitted if not a timeout or ban
}

// Singular message deletion
type ClearMsgMessage struct {
	Channel string      `json:"channel"`
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"` // the deleted message
	Type    MessageType `json:"type"`

	Login       string `json:"login"`         // name of user who sent deleted message
	TargetMsgID string `json:"target-msg-id"` // msg id of the deleted message
}

// Information about user that successfully logged in
type GlobalUserstateMessage struct {
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`

	EmoteSets []string `json:"emote-sets"` // emotes belonging to one or more emote sets
	User      *User    `json:"user"`       // information about the user that logged in
}

type HostTargetMessage struct {
	Channel string      `json:"channel"`
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"` // "<channel> began hosting <hosted>" or "<channel> exited host mode"
	Type    MessageType `json:"type"`

	Hosted  string `json:"hosted"`  // the channel that was hosted by channel
	Viewers int    `json:"viewers"` // the number of viewers at channel during the host event
}

type NoticeMessage struct {
	Channel string      `json:"channel"`
	Data    *IRCData    `json:"data"`
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
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`
}

type RoomstateMessage struct {
	Channel string      `json:"channel"`
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`

	States []RoomState `json:"states"`  // the states in the roomstate tags
	RoomID string      `json:"room-id"` // channel ID
}

type RoomState struct { // note followers-only: -1 (disabled), 0 (enabled immediate chat), > 0 (enabled number of minutes delay to chat)
	Mode    string `json:"mode"`    // emote-only, followers-only, uniquechat(r9k), slowmode(slow), subs-only
	Enabled bool   `json:"enabled"` // mode turned on or off
	Delay   int    `json:"delay"`   // seconds between messages (slow), minutes post-follow (followers-only)
}

type UsernoticeMessage struct {
	Channel string      `json:"channel"`
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`

	MsgParams map[string]string `json:"msg-params"` // any msg-param tags for the notice
	SystemMsg string            `json:"system-msg"` // message printed in chat on the notice
	User      *User             `json:"user"`       // user who caused the notice
}

type UserstateMessage struct {
	Channel string      `json:"channel"`
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`

	EmoteSets []string `json:"emote-sets"` // emotes belonging to one or more emote sets
	User      *User    `json:"user"`       // user that joined or sent a privmsg
}

type NamesMessage struct { // warning: technically deprecated, but not quite removed yet
	Channel string      `json:"channel"`
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`

	Users []string `json:"users"` // list of usernames
}

type JoinMessage struct {
	Channel string      `json:"channel"`
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`

	Username string `json:"username"` // name of joined account
}

type PartMessage struct {
	Channel string      `json:"channel"`
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Type    MessageType `json:"type"`

	Username string `json:"username"` // name of parted account
}

type PingMessage struct {
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
}

type PongMessage struct {
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
}

type PrivmsgMessage struct {
	Channel string      `json:"channel"`
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`

	Action bool  `json:"action"` // indicates if the /me command was used
	User   *User `json:"user"`   // user that sent the message
}

type WhisperMessage struct {
	Channel string      `json:"channel"`
	Data    *IRCData    `json:"data"`
	IRCType string      `json:"irc-type"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`

	Action bool   `json:"action"` // indicates if the /me command was used
	Target string `json:"target"` // message recipient
	User   *User  `json:"user"`   // message sender
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
	End   int `json:"end"`
	Start int `json:"start"`
}

type User struct {
	BadgeInfo    string   `json:"badgeinfo"`
	Badges       []*Badge `json:"badges"`
	Bits         int      `json:"bits"`
	Broadcaster  bool     `json:"broadcaster"`
	Color        string   `json:"color"`
	DisplayName  string   `json:"displayname"`
	Emotes       []*Emote `json:"emotes"`
	Flags        string   `json:"flags"`
	ID           string   `json:"id"`
	Mod          bool     `json:"mod"`
	Name         string   `json:"name"`
	RoomID       string   `json:"roomid"`
	Subscriber   bool     `json:"subscriber"`
	TmiSentTs    string   `json:"tmisentts"`
	Turbo        bool     `json:"turbo"`
	UserID       string   `json:"userid"`
	UserType     string   `json:"usertype"`
	BadgeInfoRaw string   `json:"badgeinforaw"`
	BadgesRaw    string   `json:"badgesraw"`
	EmotesRaw    string   `json:"emotesraw"`
}

func (msg *UnsetMessage) GetType() MessageType {
	return msg.Type
}
func (msg *WelcomeMessage) GetType() MessageType {
	return msg.Type
}
func (msg *InvalidIRCMessage) GetType() MessageType {
	return msg.Type
}
func (msg *ClearChatMessage) GetType() MessageType {
	return msg.Type
}
func (msg *ClearMsgMessage) GetType() MessageType {
	return msg.Type
}
func (msg *GlobalUserstateMessage) GetType() MessageType {
	return msg.Type
}
func (msg *HostTargetMessage) GetType() MessageType {
	return msg.Type
}
func (msg *NoticeMessage) GetType() MessageType {
	return msg.Type
}
func (msg *ReconnectMessage) GetType() MessageType {
	return msg.Type
}
func (msg *RoomstateMessage) GetType() MessageType {
	return msg.Type
}
func (msg *UsernoticeMessage) GetType() MessageType {
	return msg.Type
}
func (msg *UserstateMessage) GetType() MessageType {
	return msg.Type
}
func (msg *NamesMessage) GetType() MessageType {
	return msg.Type
}
func (msg *JoinMessage) GetType() MessageType {
	return msg.Type
}
func (msg *PartMessage) GetType() MessageType {
	return msg.Type
}
func (msg *PingMessage) GetType() MessageType {
	return msg.Type
}
func (msg *PongMessage) GetType() MessageType {
	return msg.Type
}
func (msg *PrivmsgMessage) GetType() MessageType {
	return msg.Type
}
func (msg *WhisperMessage) GetType() MessageType {
	return msg.Type
}