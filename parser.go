package tmi

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func (tags IRCTags) EscapeIRCTagValues() {
	var ircTagEscapes = []struct {
		from string
		to   string
	}{
		{`\s`, ` `},
		{`\n`, ``},
		{`\r`, ``},
		{`\:`, `;`},
		{`\\`, `\`},
	}

	for k, v := range tags {
		for _, escape := range ircTagEscapes {
			v = strings.ReplaceAll(v, escape.from, escape.to)
		}
		v = strings.TrimSuffix(v, "\\")
		v = strings.TrimSpace(v)
		tags[k] = v
	}
}

func parseIRCMessage(message string) (IRCData, error) {
	ircData := IRCData{
		Raw:    message,
		Params: []string{},
	}

	fields := strings.Fields(message)
	if len(fields) == 0 {
		return ircData, errors.New("parseIRCMessage: empty")
	}
	var idx int

	if strings.HasPrefix(fields[idx], "@") {
		ircData.Tags = parseTags(fields[idx])
		idx++
	}

	if idx == len(fields) {
		return ircData, errors.New("parseIRCMessage: only tags")
	}

	if strings.HasPrefix(fields[idx], ":") {
		ircData.Prefix = strings.TrimPrefix(fields[idx], ":")
		idx++
	}

	if idx == len(fields) {
		return ircData, errors.New("parseIRCMessage: no command")
	}

	ircData.Command = fields[idx]
	idx++

	if idx == len(fields) {
		return ircData, nil
	}

	var msgIdx = -1
	for i, v := range fields[idx:] {
		if strings.HasPrefix(v, ":") {
			msgIdx = idx + i
			break
		}
	}
	if msgIdx >= 0 {
		ircData.Params = fields[idx:msgIdx]
		var msgSlice = fields[msgIdx:]
		msgSlice[0] = strings.TrimPrefix(msgSlice[0], ":")
		var message = strings.Join(msgSlice, " ")
		ircData.Params = append(ircData.Params, message)
	} else {
		ircData.Params = fields[idx:]
	}

	return ircData, nil
}

func parseTags(rawTags string) IRCTags {
	var tags IRCTags = make(map[string]string)

	rawTags = strings.TrimPrefix(rawTags, "@")
	splRawTags := strings.Split(rawTags, ";")

	for _, tag := range splRawTags {
		pair := strings.SplitN(tag, "=", 2)

		var key string = pair[0]
		var val string
		if len(pair) == 2 {
			val = pair[1]
		}

		tags[key] = val
	}

	return tags
}

func parseUnsetMessage(ircData IRCData) UnsetMessage {
	return UnsetMessage{
		Data:    ircData,
		IRCType: ircData.Command,
		Text:    ircData.Raw,
		Type:    UNSET,
	}
}

func parseClearChatMessage(data IRCData) ClearChatMessage {
	var clearChatMessage = ClearChatMessage{
		Data:    data,
		IRCType: data.Command,
		Type:    CLEARCHAT,
	}

	var bAlloc int // for growing string builder

	clearChatMessage.Channel = strings.TrimPrefix(data.Params[0], "#")
	bAlloc += len(clearChatMessage.Channel)

	if len(data.Params) == 2 {
		bAlloc += 27 // " was permanently banned in " or " timed out for {banDuration} seconds in "

		clearChatMessage.Target = data.Params[1]
		bAlloc += len(clearChatMessage.Target)

		if banDuration, ok := data.Tags["ban-duration"]; ok {
			bAlloc += len(banDuration)
			if duration, err := strconv.Atoi(banDuration); err == nil {
				clearChatMessage.BanDuration = time.Duration(duration) * time.Second
			}
		} else {
			clearChatMessage.BanDuration = -1
		}
	} else {
		bAlloc += 16 // "chat cleared in "
		clearChatMessage.BanDuration = -1
	}

	var b strings.Builder
	b.Grow(bAlloc)

	if clearChatMessage.Target == "" {
		b.WriteString("chat cleared in ")
	} else {
		b.WriteString(clearChatMessage.Target)
		if clearChatMessage.BanDuration < 0 {
			b.WriteString(" was permanently banned in ")
		} else {
			b.WriteString(" timed out for ")
			b.WriteString(data.Tags["ban-duration"])
			b.WriteString(" seconds in ")
		}
	}
	b.WriteString(clearChatMessage.Channel)
	clearChatMessage.Text = b.String()

	return clearChatMessage
}

func parseClearMsgMessage(data IRCData) ClearMsgMessage {
	var clearMsgMessage = ClearMsgMessage{
		Data:        data,
		IRCType:     data.Command,
		Type:        CLEARMSG,
		Login:       data.Tags["login"],
		TargetMsgID: data.Tags["target-msg-id"],
	}

	clearMsgMessage.Channel = strings.TrimPrefix(data.Params[0], "#")

	if len(data.Params) == 2 {
		clearMsgMessage.Text = data.Params[1]
	}

	return clearMsgMessage
}

func parseGlobalUserstateMessage(data IRCData) GlobalUserstateMessage {
	var globalUserstateMessage = GlobalUserstateMessage{
		Data:      data,
		IRCType:   data.Command,
		Type:      GLOBALUSERSTATE,
		EmoteSets: parseEmoteSets(data.Tags),
		User:      parseUser(data.Tags, data.Prefix),
	}

	return globalUserstateMessage
}

func parseEmoteSets(tags IRCTags) []string {
	if sets, ok := tags["emote-sets"]; ok {
		return strings.Split(sets, ",")
	} else {
		return []string{}
	}
}

func parseUser(tags IRCTags, prefix string) *User {
	var user = User{
		BadgeInfo:   tags["badge-info"],
		Color:       tags["color"],
		DisplayName: tags["display-name"],
		Mod:         tags["mod"] == "1",
		RoomID:      tags["room-id"],
		Subscriber:  tags["subscriber"] == "1",
		TmiSentTs:   tags["tmi-sent-ts"],
		Turbo:       tags["turbo"] == "1",
		UserID:      tags["user-id"],
		UserType:    tags["user-type"],
		BadgesRaw:   tags["badges"],
	}

	if bits, ok := tags["bits"]; ok {
		if val, err := strconv.Atoi(bits); err == nil {
			user.Bits = val
		}
	}

	if user.DisplayName != "" {
		user.Name = strings.ToLower(user.DisplayName)
	} else {
		// TODO:
		// parsePrefix
	}

	user.Badges = parseBadges(user.BadgesRaw)

	if len(user.Badges) > 0 {
		for _, badge := range user.Badges {
			if badge.Name == "broadcaster" {
				user.Broadcaster = true
			}
		}
	}

	return &user
}

func parseEmotes(rawEmotes, message string) []Emote {
	var emotes []Emote
	if rawEmotes == "" {
		return emotes
	}

	msg := []rune(message)

	var splEmotes = strings.Split(rawEmotes, "/")

parseLoop:
	for _, emote := range splEmotes {
		var spl = strings.SplitN(emote, ":", 2)

		var posPairs = strings.Split(spl[1], ",")
		if len(posPairs) < 1 {
			continue
		}

		var positions = []EmotePosition{}
		for _, pair := range posPairs {
			var position = strings.SplitN(pair, "-", 2)
			if len(position) != 2 {
				continue parseLoop
			}
			var startIdx, endIdx int
			var err error

			startIdx, err = strconv.Atoi(position[0])
			if err != nil {
				continue parseLoop
			}

			endIdx, err = strconv.Atoi(position[1])
			if err != nil {
				continue parseLoop
			}

			positions = append(positions, EmotePosition{
				StartIdx: startIdx,
				EndIdx:   endIdx,
			})
		}

		var nameStartIdx = positions[0].StartIdx
		if nameStartIdx+1 > len(msg) {
			nameStartIdx = len(msg) - 1
		}
		var nameEndIdx = positions[0].EndIdx
		if nameEndIdx+1 > len(msg) {
			nameEndIdx = len(msg) - 1
		}

		emotes = append(emotes, Emote{
			ID:        spl[0],
			Name:      string(msg[nameStartIdx : nameEndIdx+1]),
			Positions: positions,
		})
	}

	return emotes
}

func parseBadges(rawBadges string) []Badge {
	var badges []Badge
	if rawBadges == "" {
		return badges
	}

	var splBadges = strings.Split(rawBadges, ",")

	for _, b := range splBadges {
		var pair = strings.SplitN(b, "/", 2)
		var badge Badge
		badge.Name = pair[0]
		if val, err := strconv.Atoi(pair[1]); err == nil {
			badge.Value = val
		}
		badges = append(badges, badge)
	}

	return badges
}

func parseNoticeMessage(data IRCData) (NoticeMessage, error) {
	var noticeMessage = NoticeMessage{
		Data:    data,
		IRCType: data.Command,
		Notice:  "notice",
		Type:    NOTICE,
	}
	noticeMessage.Channel = strings.TrimPrefix(data.Params[0], "#")
	var msg string
	if len(data.Params) == 2 {
		msg = data.Params[1]
		noticeMessage.Text = msg
	}
	if msgId, ok := data.Tags["msg-id"]; ok {
		noticeMessage.MsgID = msgId

		switch msgId {
		// Automod
		case "msg_rejected",
			"msg_rejected_mandatory":
			noticeMessage.Notice = "automod"

		// Emote only mode on/off
		case "emote_only_off":
			noticeMessage.Notice = "emoteonly"
			noticeMessage.Enabled = false
		case "emote_only_on":
			noticeMessage.Notice = "emoteonly"
			noticeMessage.Enabled = true

		// Moderators of the channel, or none
		case "no_mods":
			noticeMessage.Notice = "mods"
		case "room_mods":
			noticeMessage.Notice = "mods"
			noticeMessage.Mods = []string{}
			var modStr = msg
			var mods = strings.Split(strings.ToLower(strings.Split(modStr, ": ")[1]), ", ")
			for _, v := range mods {
				if v != "" {
					noticeMessage.Mods = append(noticeMessage.Mods, v)
				}
			}

		// r9k (uniquechat) mode on/off
		case "r9k_off":
			noticeMessage.Notice = "uniquechat"
			noticeMessage.Enabled = false
		case "r9k_on":
			noticeMessage.Notice = "uniquechat"
			noticeMessage.Enabled = true

		// Subscribers only mode on/off
		case "subs_off":
			noticeMessage.Notice = "subonly"
			noticeMessage.Enabled = false
		case "subs_on":
			noticeMessage.Notice = "subonly"
			noticeMessage.Enabled = true

		// VIPs of the channel, or none
		case "no_vips":
			noticeMessage.Notice = "vips"
		case "vips_success":
			noticeMessage.Notice = "vips"
			noticeMessage.VIPs = []string{}
			var vipStr = msg
			vipStr = strings.TrimSuffix(vipStr, ".")
			var vips = strings.Split(strings.ToLower(strings.Split(vipStr, ": ")[1]), ", ")
			for _, v := range vips {
				if v != "" {
					noticeMessage.VIPs = append(noticeMessage.VIPs, v)
				}
			}

		// Listen for ROOMSTATE followers notice instead (includes delay)
		case "followers_off":
		case "followers_on":
		case "followers_onzero":

		// Listen for HOSTTARGET instead
		case "host_off":
		case "host_on":

		// Listen for ROOMSTATE slowmode notice instead (includes delay)
		case "slow_off":
		case "slow_on":
		}
	} else {
		loginFailures := []string{
			"Login unsuccessful",
			"Login authentication failed",
			"Error logging in",
			"Improperly formatted auth",
			"Invalid NICK",
		}
		for _, failure := range loginFailures {
			if strings.Contains(msg, failure) {
				noticeMessage.MsgID = "login_failure"
				return noticeMessage, errors.New(msg)
			}
		}
		noticeMessage.MsgID = "parse_error"
		return noticeMessage, errors.New("could not properly parse NOTICE:\n" + data.Raw)
	}

	return noticeMessage, nil
}
