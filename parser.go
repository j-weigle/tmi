package tmi

import (
	"strings"
)

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

type Message struct {
	From      string     `json:"from"`
	Text      string     `json:"text"`
	Userstate *Userstate `json:"userstate"`
}

type MessageData struct {
	Raw     string            `json:"raw"`
	Tags    map[string]string `json:"tags"`
	Prefix  string            `json:"prefix"`
	Command string            `json:"command"`
	Params  []string          `json:"params"`
}

type Userstate struct {
	//TODO: Bits
	BadgeInfo    string   `json:"badgeinfo"`
	Badges       []*Badge `json:"badges"`
	Color        string   `json:"color"`
	DisplayName  string   `json:"displayname"`
	Emotes       []*Emote `json:"emotes"`
	Flags        string   `json:"flags"`
	Id           string   `json:"id"`
	Mod          bool     `json:"mod"`
	RoomId       string   `json:"roomid"`
	Subscriber   bool     `json:"subscriber"`
	TmiSentTs    string   `json:"tmisentts"`
	Turbo        bool     `json:"turbo"`
	UserId       string   `json:"userid"`
	Username     string   `json:"username"`
	UserType     string   `json:"usertype"`
	BadgeInfoRaw string   `json:"badgeinforaw"`
	BadgesRaw    string   `json:"badgesraw"`
	EmotesRaw    string   `json:"emotesraw"`
	MessageType  string   `json:"messagetype"`
}

func parseMessage(message string) *MessageData {
	msgdata := &MessageData{
		Raw: message,
	}

	splMessage := strings.Fields(message)
	if len(splMessage) == 0 {
		return msgdata
	}

	if splMessage[0][0] == '@' {
		msgdata.Tags = parseTags(splMessage[0][1:])
		msgdata.Prefix = splMessage[1][1:]
		msgdata.Command = splMessage[2]

		if len(splMessage) > 3 {
			msgdata.Params = parseParams(splMessage[3:])
		}
	} else if splMessage[0][0] == ':' {
		msgdata.Prefix = splMessage[0][1:]
		msgdata.Command = splMessage[1]

		if len(splMessage) > 2 {
			msgdata.Params = parseParams(splMessage[2:])
		}
	} else {
		msgdata.Command = splMessage[0]

		if len(splMessage) > 1 {
			msgdata.Params = parseParams(splMessage[1:])
		}
	}

	return msgdata
}

func parseParams(rawParams []string) []string {
	var params []string
	var msgIdx int
	var msg string

	for i := range rawParams {
		if rawParams[i][0] == ':' {
			msgIdx = i
			break
		}
	}

	if msgIdx != 0 {
		msgSlice := rawParams[msgIdx:]
		msgSlice[0] = msgSlice[0][1:]
		msg = strings.Join(msgSlice, " ")
	}

	if msg != "" {
		rawParams[msgIdx] = msg
		params = rawParams[:msgIdx+1]
	} else {
		params = rawParams
	}

	return params
}

func parseTags(rawTags string) map[string]string {
	tags := make(map[string]string)

	splRawTags := strings.Split(rawTags, ";")

	for _, tag := range splRawTags {
		pair := strings.Split(tag, "=")
		if strings.Contains(tag, "=") {
			tags[pair[0]] = pair[1]
		} else {
			tags[pair[0]] = ""
		}
	}

	return tags
}
