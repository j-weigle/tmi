package tmi

import (
	"strings"
	"testing"
)

func (d1 *IRCData) equals(d2 *IRCData) bool {
	if d1.Raw != d2.Raw ||
		d1.Prefix != d2.Prefix ||
		d1.Command != d2.Command {
		return false
	}
	for k, v1 := range d1.Tags {
		if v2, ok := d2.Tags[k]; ok {
			if v1 != v2 {
				return false
			}
		} else {
			return false
		}
	}
	if len(d1.Params) != len(d2.Params) {
		return false
	}
	for i, v := range d1.Params {
		if v != d2.Params[i] {
			return false
		}
	}
	return true
}

type testingIRCParseData struct {
	tags    []tag
	prefix  string
	command string
	params  []string
}

type tag struct {
	key   string
	value string
}

func getRawString(data testingIRCParseData) string {
	var rawStringSlice = []string{}
	// Tags
	if data.tags != nil {
		var rawTags string
		var tags = make([]string, len(data.tags))
		for _, tag := range data.tags {
			tags = append(tags, tag.key+"="+tag.value)
		}
		rawTags = "@" + strings.Join(tags, ";")
		rawStringSlice = append(rawStringSlice, rawTags)
	}
	// Prefix
	if data.prefix != "" {
		data.prefix = ":" + data.prefix
		rawStringSlice = append(rawStringSlice, data.prefix)
	}
	// Command
	if data.command != "" {
		rawStringSlice = append(rawStringSlice, data.command)
	}
	// Params
	if data.params != nil && len(data.params) != 0 {
		rawStringSlice = append(rawStringSlice, strings.Join(data.params, " "))
	}

	return strings.Join(rawStringSlice, " ")
}

func getIRCData(data testingIRCParseData) *IRCData {
	var tags = make(map[string]string)
	var params []string
	for _, t := range data.tags {
		tags[t.key] = t.value
	}
	return &IRCData{
		Raw:     getRawString(data),
		Tags:    tags,
		Prefix:  data.prefix,
		Command: data.command,
		Params:  params,
	}
}

func TestParseIRCMessage(t *testing.T) {
	// TODO:
}
