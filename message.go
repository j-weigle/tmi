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
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type InvalidIRCMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type ClearChatMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type ClearMsgMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type GlobalUserstateMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type HostTargetMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type NoticeMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`

	Mods   []string `json:"mods"`
	MsgId  string   `json:"msg-id"`
	Notice string   `json:"notice"`
	On     bool     `json:"on"`
	Vips   []string `json:"vips"`
}

type ReconnectMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type RoomstateMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type UsernoticeMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type UserstateMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type ModeMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type NamesMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type JoinMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type PartMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type PingMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type PongMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type PrivmsgMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`

	// TODO:
	Action bool  `json:"action"`
	User   *User `json:"user"`
}

type WhisperMessage struct {
	From    string      `json:"from"`
	IRCType string      `json:"irc-type"`
	RawIRC  *IRCData    `json:"raw-irc"`
	Text    string      `json:"text"`
	Type    MessageType `json:"type"`
	Error   error       `json:"error"`
	// TODO:
}

type Badge struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type Emote struct {
	Id    string `json:"id"`
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
	Id           string   `json:"id"`
	Mod          bool     `json:"mod"`
	Name         string   `json:"name"`
	RoomId       string   `json:"roomid"`
	Subscriber   bool     `json:"subscriber"`
	TmiSentTs    string   `json:"tmisentts"`
	Turbo        bool     `json:"turbo"`
	UserId       string   `json:"userid"`
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
