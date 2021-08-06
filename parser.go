package tmi

import (
	"fmt"
	"strings"
)

// TODO: make IRC parser more robust

func parseIRCMessage(message string) *IRCData {
	ircData := &IRCData{
		Raw: message,
	}

	fields := strings.Fields(message)
	if len(fields) == 0 {
		return ircData
	}

	if strings.HasPrefix(fields[0], "@") {
		ircData.Tags = parseTags(fields[0][1:])
		ircData.Prefix = fields[1][1:]
		ircData.Command = fields[2]

		if len(fields) > 3 {
			ircData.Params = parseParams(fields[3:])
		}
	} else if strings.HasPrefix(fields[0], ":") {
		ircData.Prefix = fields[0][1:]
		ircData.Command = fields[1]

		if len(fields) > 2 {
			ircData.Params = parseParams(fields[2:])
		}
	} else {
		ircData.Command = fields[0]

		if len(fields) > 1 {
			ircData.Params = parseParams(fields[1:])
		}
	}

	return ircData
}

func parseNoticeMessage(ircData *IRCData) (*NoticeMessage, error) {
	var noticeMessage = &NoticeMessage{
		IRCType: ircData.Command,
		RawIRC:  ircData,
		Type:    NOTICE,
		Notice:  "notice",
	}
	if len(ircData.Params) >= 1 {
		noticeMessage.From = strings.TrimPrefix(ircData.Params[0], "#")
	}
	var msg string
	if len(ircData.Params) >= 2 {
		msg = ircData.Params[1]
		noticeMessage.Text = msg
	}
	if msgId, ok := ircData.Tags["msg-id"]; ok {
		noticeMessage.MsgId = msgId

		switch msgId {
		// Automod
		case "msg_rejected",
			"msg_rejected_mandatory":
			noticeMessage.Notice = "automod"

		// Emote only mode on/off
		case "emote_only_off":
			noticeMessage.Notice = "emoteonly"
			noticeMessage.On = false
		case "emote_only_on":
			noticeMessage.Notice = "emoteonly"
			noticeMessage.On = true

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
			noticeMessage.On = false
		case "r9k_on":
			noticeMessage.Notice = "uniquechat"
			noticeMessage.On = true

		// Subscribers only mode on/off
		case "subs_off":
			noticeMessage.Notice = "subonly"
			noticeMessage.On = false
		case "subs_on":
			noticeMessage.Notice = "subonly"
			noticeMessage.On = true

		// VIPs of the channel, or none
		case "no_vips":
			noticeMessage.Notice = "vips"
		case "vips_success":
			noticeMessage.Notice = "vips"
			noticeMessage.Vips = []string{}
			var vipStr = msg
			vipStr = strings.TrimSuffix(vipStr, ".")
			var vips = strings.Split(strings.ToLower(strings.Split(vipStr, ": ")[1]), ", ")
			for _, v := range vips {
				if v != "" {
					noticeMessage.Vips = append(noticeMessage.Vips, v)
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
				noticeMessage.MsgId = "login-failure"
				return noticeMessage, fmt.Errorf("login authentication\n" + msg)
			}
		}
		noticeMessage.MsgId = "parse-error"
		return noticeMessage, fmt.Errorf("could not parse NOTICE:\n" + ircData.Raw)
	}

	return noticeMessage, nil
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
