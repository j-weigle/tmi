package tmi

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

type IRCData struct {
	Raw     string            `json:"raw"`
	Tags    map[string]string `json:"tags"`
	Prefix  string            `json:"prefix"`
	Command string            `json:"command"`
	Params  []string          `json:"params"`
}

type Message struct {
	Action  bool         `json:"action"`
	From    string       `json:"from"`
	Info    *MessageInfo `json:"info"`
	IRCType string       `json:"irc-type"`
	MsgId   string       `json:"msg-id"`
	Type    string       `json:"type"`
	RawIRC  *IRCData     `json:"raw-irc"`
	Text    string       `json:"text"`
	User    *User        `json:"user"`
}

type MessageInfo struct {
	Mods []string `json:"mods"`
	On   bool     `json:"on"`
	Vips []string `json:"vips"`
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
