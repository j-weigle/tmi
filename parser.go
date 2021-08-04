package tmi

import (
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
