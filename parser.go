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

func ParseTimeStamp(unixTime string) time.Time {
	var i, err = strconv.ParseInt(unixTime, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(0, i*int64(time.Millisecond))
}

func ParseReplyParentMessage(tags IRCTags) ReplyMsgParent {
	return ReplyMsgParent{
		DisplayName: tags["reply-parent-display-name"],
		ID:          tags["reply-parent-msg-id"],
		Text:        tags["reply-parent-msg-body"],
		UserID:      tags["reply-parent-user-id"],
		Username:    tags["reply-parent-user-login"],
	}
}

func parseIRCMessage(message string) (IRCData, error) {
	ircData := IRCData{
		Raw:    message,
		Params: []string{},
		Tags:   make(IRCTags),
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
	if len(data.Params) > 0 {
		clearChatMessage.Channel = data.Params[0]
	}

	// for growing string builder
	var bAlloc = len(clearChatMessage.Channel)

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
	if len(data.Params) > 0 {
		clearMsgMessage.Channel = data.Params[0]
	}

	if len(data.Params) == 2 {
		clearMsgMessage.Text = data.Params[1]
	}

	return clearMsgMessage
}

func parseGlobalUserstateMessage(data IRCData) GlobalUserstateMessage {
	return GlobalUserstateMessage{
		Data:      data,
		IRCType:   data.Command,
		Type:      GLOBALUSERSTATE,
		EmoteSets: parseEmoteSets(data.Tags),
		User:      parseUser(data.Tags, data.Prefix),
	}
}

func parseHostTargetMessage(data IRCData) HostTargetMessage {
	var hostTargetMessage = HostTargetMessage{
		Data:    data,
		IRCType: data.Command,
		Type:    HOSTTARGET,
	}
	if len(data.Params) > 0 {
		hostTargetMessage.Channel = data.Params[0]
	}

	// for growing string builder
	var bAlloc = len(hostTargetMessage.Channel)

	var viewers string
	if len(data.Params) == 2 {
		var fields = strings.Fields(data.Params[1])
		if fields[0] == "-" {
			bAlloc += 17 // " exited host mode"
		} else {
			hostTargetMessage.Hosted = fields[0]
			bAlloc += len(hostTargetMessage.Hosted) // " is now hosting {hosted}"
		}
		if len(fields) == 2 {
			viewers = fields[1]
			if v, err := strconv.Atoi(viewers); err == nil {
				hostTargetMessage.Viewers = v
			}
			bAlloc += len(viewers) + 14 // " with {viewers} viewers"
		}
	}

	var b strings.Builder
	b.Grow(bAlloc)

	b.WriteString(hostTargetMessage.Channel)
	if hostTargetMessage.Hosted != "" {
		b.WriteString(" is now hosting ")
		b.WriteString(hostTargetMessage.Hosted)

	} else {
		b.WriteString(" exited host mode")
	}
	if viewers != "" {
		b.WriteString(" with ")
		b.WriteString(viewers)
		b.WriteString(" viewers")
	}
	hostTargetMessage.Text = b.String()

	return hostTargetMessage
}

func parseNoticeMessage(data IRCData) (NoticeMessage, error) {
	var noticeMessage = NoticeMessage{
		Data:    data,
		IRCType: data.Command,
		Notice:  "notice",
		Type:    NOTICE,
	}
	if len(data.Params) > 0 {
		noticeMessage.Channel = data.Params[0]
	}

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
			// There are no moderators for this room.
			noticeMessage.Notice = "mods"
		case "room_mods":
			// The moderators of this room are: mod1, mod2, etc
			noticeMessage.Notice = "mods"
			var splMsg = strings.Split(msg, ": ")
			if len(splMsg) == 2 {
				var mods = strings.ToLower(splMsg[1])
				var modList = strings.Split(mods, ", ")
				for _, mod := range modList {
					if mod != "" {
						noticeMessage.Mods = append(noticeMessage.Mods, mod)
					}
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
			// There are no VIPs for this room.
			noticeMessage.Notice = "vips"
		case "vips_success":
			// The VIPs of this room are: VIP1, VIP2, etc
			noticeMessage.Notice = "vips"
			var splMsg = strings.Split(msg, ": ")
			if len(splMsg) == 2 {
				var vips = strings.ToLower(splMsg[1])
				var vipList = strings.Split(vips, ", ")
				for _, vip := range vipList {
					if vip != "" {
						noticeMessage.VIPs = append(noticeMessage.VIPs, vip)
					}
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
				return noticeMessage, ErrLoginFailure
			}
		}
		noticeMessage.MsgID = "parse_error"
	}

	return noticeMessage, nil
}

func parseReconnectMessage(data IRCData) ReconnectMessage {
	return ReconnectMessage{
		Data:    data,
		IRCType: data.Command,
		Type:    RECONNECT,
	}
}

func parseRoomstateMessage(data IRCData) RoomstateMessage {
	var roomstateMessage = RoomstateMessage{
		Data:    data,
		IRCType: data.Command,
		Type:    ROOMSTATE,
		States:  make(map[string]RoomState),
	}
	if len(data.Params) > 0 {
		roomstateMessage.Channel = data.Params[0]
	}

	var modeTags = [6]string{
		"emote-only",
		"followers-only",
		"r9k",
		"rituals",
		"slow",
		"subs-only"}
	for _, tag := range modeTags {
		if v, ok := data.Tags[tag]; ok {
			var roomstate = RoomState{}
			if val, err := strconv.Atoi(v); err == nil {
				switch tag {
				case "followers-only":
					switch val {
					case -1:
					case 0:
						roomstate.Enabled = true
					default:
						roomstate.Enabled = true
						roomstate.Delay = time.Duration(val) * time.Minute
					}
				case "slow":
					if val != 0 {
						roomstate.Enabled = true
						roomstate.Delay = time.Duration(val) * time.Second
					}
				default:
					roomstate.Enabled = val == 1
				}
			}

			roomstateMessage.States[tag] = roomstate
		}
	}

	return roomstateMessage
}

func parseUsernoticeMessage(data IRCData) UsernoticeMessage {
	var usernoticeMessage = UsernoticeMessage{
		Data:      data,
		IRCType:   data.Command,
		Type:      USERNOTICE,
		MsgParams: make(IRCTags),
		SystemMsg: data.Tags["system-msg"],
		User:      parseUser(data.Tags, data.Prefix),
	}
	if len(data.Params) > 0 {
		usernoticeMessage.Channel = data.Params[0]
	}

	if len(data.Params) == 2 {
		usernoticeMessage.Text = data.Params[1]
	}

	usernoticeMessage.Emotes = parseEmotes(data.Tags["emotes"], usernoticeMessage.Text)

	for t, v := range data.Tags {
		if strings.HasPrefix(t, "msg-param") {
			usernoticeMessage.MsgParams[t] = v
		}
	}

	return usernoticeMessage
}

func parseUserstateMessage(data IRCData) UserstateMessage {
	var userstateMessage = UserstateMessage{
		Data:      data,
		IRCType:   data.Command,
		Type:      USERSTATE,
		EmoteSets: parseEmoteSets(data.Tags),
		User:      parseUser(data.Tags, data.Prefix),
	}
	if len(data.Params) > 0 {
		userstateMessage.Channel = data.Params[0]
	}
	return userstateMessage
}

func parseNamesMessage(data IRCData) NamesMessage {
	// WARNING: deprecated, but not removed yet
	var namesMessage = NamesMessage{
		Data:    data,
		IRCType: data.Command,
		Type:    NAMES,
	}

	if len(data.Params) == 4 {
		namesMessage.Channel = data.Params[2]
		namesMessage.Users = strings.Fields(data.Params[3])
	}

	return namesMessage
}

func parseJoinMessage(data IRCData) JoinMessage {
	var joinMessage = JoinMessage{
		Data:     data,
		IRCType:  data.Command,
		Type:     JOIN,
		Username: parseUsernameFromPrefix(data.Prefix),
	}
	if len(data.Params) > 0 {
		joinMessage.Channel = data.Params[0]
	}
	return joinMessage
}

func parsePartMessage(data IRCData) PartMessage {
	var partMessage = PartMessage{
		Data:     data,
		IRCType:  data.Command,
		Type:     PART,
		Username: parseUsernameFromPrefix(data.Prefix),
	}
	if len(data.Params) > 0 {
		partMessage.Channel = data.Params[0]
	}
	return partMessage
}

func parsePingMessage(data IRCData) PingMessage {
	var pingMessage = PingMessage{
		Data:    data,
		IRCType: data.Command,
		Type:    PING,
	}
	if len(data.Params) == 1 {
		pingMessage.Text = data.Params[0]
	}
	return pingMessage
}

func parsePongMessage(data IRCData) PongMessage {
	var pongMessage = PongMessage{
		Data:    data,
		IRCType: data.Command,
		Type:    PONG,
	}
	if len(data.Params) == 2 {
		pongMessage.Text = data.Params[1]
	}
	return pongMessage
}

func parsePrivateMessage(data IRCData) PrivateMessage {
	var privateMessage = PrivateMessage{
		Data:    data,
		IRCType: data.Command,
		Type:    PRIVMSG,
		ID:      data.Tags["id"],
		User:    parseUser(data.Tags, data.Prefix),
	}
	if len(data.Params) > 0 {
		privateMessage.Channel = data.Params[0]
	}

	if len(data.Params) == 2 {
		privateMessage.Text = data.Params[1]
	}

	if strings.HasPrefix(privateMessage.Text, "\u0001ACTION") &&
		strings.HasSuffix(privateMessage.Text, "\u0001") {
		privateMessage.Text = privateMessage.Text[len("\u0001ACTION ") : len(privateMessage.Text)-1]
		privateMessage.Action = true
	}

	privateMessage.Emotes = parseEmotes(data.Tags["emotes"], privateMessage.Text)

	if bits, ok := data.Tags["bits"]; ok {
		if val, err := strconv.Atoi(bits); err == nil {
			privateMessage.Bits = val
		}
	}

	if _, ok := data.Tags["reply-parent-msg-id"]; ok {
		privateMessage.Reply = true
	}

	return privateMessage
}

func parseWhisperMessage(data IRCData) WhisperMessage {
	var whisperMessage = WhisperMessage{
		Data:    data,
		IRCType: data.Command,
		Type:    WHISPER,
		ID:      data.Tags["message-id"],
		User:    parseUser(data.Tags, data.Prefix),
	}
	if len(data.Params) > 0 {
		whisperMessage.Target = data.Params[0]
	}

	if len(data.Params) == 2 {
		whisperMessage.Text = data.Params[1]
	}

	if strings.HasPrefix(whisperMessage.Text, "/me") {
		whisperMessage.Text = whisperMessage.Text[len("/me "):]
		whisperMessage.Action = true
	}

	whisperMessage.Emotes = parseEmotes(data.Tags["emotes"], whisperMessage.Text)

	return whisperMessage
}

func parseUser(tags IRCTags, prefix string) *User {
	var user = User{
		BadgeInfo:   tags["badge-info"],
		Color:       tags["color"],
		DisplayName: tags["display-name"],
		Mod:         tags["mod"] == "1",
		Subscriber:  tags["subscriber"] == "1",
		Turbo:       tags["turbo"] == "1",
		ID:          tags["user-id"],
		UserType:    tags["user-type"],
	}

	if user.DisplayName != "" {
		user.Name = strings.ToLower(user.DisplayName)
	} else if name, ok := tags["login"]; ok {
		user.Name = name
	} else {
		user.Name = parseUsernameFromPrefix(prefix)
	}

	user.Badges = parseBadges(tags["badges"])
	for _, badge := range user.Badges {
		if badge.Name == "broadcaster" {
			user.Broadcaster = true
		}
		if badge.Name == "vip" {
			user.VIP = true
		}
		if badge.Name == "moderator" {
			user.Mod = true
		}
		if badge.Name == "subscriber" {
			user.Subscriber = true
		}
		if badge.Name == "turbo" {
			user.Turbo = true
		}
	}

	return &user
}

func parseUsernameFromPrefix(prefix string) string {
	var username string
	if prefix != "" {
		var spl = strings.Split(prefix, "!")
		if len(spl) == 2 {
			username = spl[0]
		} else {
			spl = strings.Split(prefix, "@")
			if len(spl) == 2 {
				username = spl[0]
			}
		}
	}
	return username
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

func parseEmoteSets(tags IRCTags) []string {
	if sets, ok := tags["emote-sets"]; ok {
		return strings.Split(sets, ",")
	} else {
		return []string{}
	}
}
