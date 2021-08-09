package tmi

type MessageType int

const (
	UNSET MessageType = iota - 1
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
	MODE
	NAMES
	JOIN
	PART
	PING
	PONG
	PRIVMSG
	WHISPER
)

func (mt MessageType) String() string {
	if mt < 0 {
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
		"MODE",
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

type WelcomeMessage struct {
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Type    MessageType `json:"type"`
}

type InvalidIRCMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	// TODO:
}

type ClearChatMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	// TODO:
}

type ClearMsgMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	// TODO:
}

type GlobalUserstateMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	// TODO:
}

type HostTargetMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	// TODO:
}

type NoticeMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`

	Mods   []string `json:"mods"`   // list of mods for Channel when Notice is set to mods
	MsgID  string   `json:"msg-id"` // msg-id value from Data.Tags, or parse-error / login-error if the key doesn't exist
	Notice string   `json:"notice"` // Notice is one of: automod, emoteonly, mods, uniquechat, subonly, vips, notice (notice is the default)
	On     bool     `json:"on"`     // set when Notice is one of: emoteonly, uniquechat, subonly
	VIPs   []string `json:"vips"`   // list of vips for Channel when Notice is set to vips
}

type ReconnectMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	// TODO:
}

type RoomstateMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	// TODO:
}

type UsernoticeMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	// TODO:
}

type UserstateMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	// TODO:
}

type ModeMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	// TODO:
}

type NamesMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	// TODO:
}

type JoinMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	// TODO:
}

type PartMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	// TODO:
}

type PingMessage struct {
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
}

type PongMessage struct {
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
}

type PrivmsgMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`

	// TODO:
	Action bool  `json:"action"`
	User   *User `json:"user"`
}

type WhisperMessage struct {
	Channel string      `json:"channel"`
	IRCType string      `json:"irc-type"`
	Data    *IRCData    `json:"data"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	// TODO:
}

type Badge struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type Emote struct {
	ID    string `json:"id"`
	Start int    `json:"start"`
	End   int    `json:"end"`
	Raw   string `json:"raw"`
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
func (msg *ModeMessage) GetType() MessageType {
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
