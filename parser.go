package tmi

import (
	"errors"
	"strings"
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

func parseUnsetMessage(ircData IRCData) (UnsetMessage, error) {
	return UnsetMessage{
		Data:    ircData,
		IRCType: ircData.Command,
		Text:    ircData.Raw,
		Type:    UNSET,
	}, nil
}

func parseNoticeMessage(ircData IRCData) (NoticeMessage, error) {
	var noticeMessage = NoticeMessage{
		Data:    ircData,
		IRCType: ircData.Command,
		Notice:  "notice",
		Type:    NOTICE,
	}
	noticeMessage.Channel = strings.TrimPrefix(ircData.Params[0], "#")
	var msg string
	if len(ircData.Params) == 2 {
		msg = ircData.Params[1]
		noticeMessage.Text = msg
	}
	if msgId, ok := ircData.Tags["msg-id"]; ok {
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
		return noticeMessage, errors.New("could not properly parse NOTICE:\n" + ircData.Raw)
	}

	return noticeMessage, nil
}
