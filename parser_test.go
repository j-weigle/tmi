package tmi

import (
	"testing"
	"time"
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

func TestParseIRCMessage(t *testing.T) {
	t.Parallel()
	type parseIRCTest struct {
		in  string
		out *IRCData
	}
	var testIRCData = []*IRCData{
		{
			Raw: `@tname1=tag1;tname2=tag2;tname3=tag3 :prefix COMMAND #channel :message`,
			Tags: map[string]string{
				"tname1": "tag1",
				"tname2": "tag2",
				"tname3": "tag3",
			},
			Prefix:  "prefix",
			Command: "COMMAND",
			Params: []string{
				"#channel",
				"message",
			},
		},
		{
			Raw:     `:tmi.twitch.tv CLEARCHAT #dallas :ronni`,
			Tags:    nil,
			Prefix:  "tmi.twitch.tv",
			Command: "CLEARCHAT",
			Params: []string{
				"#dallas",
				"ronni",
			},
		},
		{
			Raw: `@login=ronni;target-msg-id=abc-123-def :tmi.twitch.tv CLEARMSG #dallas :HeyGuys`,
			Tags: map[string]string{
				"login":         "ronni",
				"target-msg-id": "abc-123-def",
			},
			Prefix:  "tmi.twitch.tv",
			Command: "CLEARMSG",
			Params: []string{
				"#dallas",
				"HeyGuys",
			},
		},
		{
			Raw: `@badge-info=subscriber/8;badges=subscriber/6;color=#0D4200;display-name=dallas;emote-sets=0,33,50,237,793,2126,3517,4578,5569,9400,10337,12239;turbo=0;user-id=1337;user-type=admin :tmi.twitch.tv GLOBALUSERSTATE`,
			Tags: map[string]string{
				"badge-info":   "subscriber/8",
				"badges":       "subscriber/6",
				"color":        "#0D4200",
				"display-name": "dallas",
				"emote-sets":   "0,33,50,237,793,2126,3517,4578,5569,9400,10337,12239",
				"turbo":        "0",
				"user-id":      "1337",
				"user-type":    "admin",
			},
			Prefix:  "tmi.twitch.tv",
			Command: "GLOBALUSERSTATE",
			Params:  []string{},
		},
		{
			Raw: `@badge-info=;badges=global_mod/1,turbo/1;color=#0D4200;display-name=ronni;emotes=25:0-4,12-16/1902:6-10;id=b34ccfc7-4977-403a-8a94-33c6bac34fb8;mod=0;room-id=1337;subscriber=0;tmi-sent-ts=1507246572675;turbo=1;user-id=1337;user-type=global_mod :ronni!ronni@ronni.tmi.twitch.tv PRIVMSG #ronni :Kappa Keepo Kappa`,
			Tags: map[string]string{
				"badge-info":   "",
				"badges":       "global_mod/1,turbo/1",
				"color":        "#0D4200",
				"display-name": "ronni",
				"emotes":       "25:0-4,12-16/1902:6-10",
				"id":           "b34ccfc7-4977-403a-8a94-33c6bac34fb8",
				"mod":          "0",
				"room-id":      "1337",
				"subscriber":   "0",
				"tmi-sent-ts":  "1507246572675",
				"turbo":        "1",
				"user-id":      "1337",
				"user-type":    "global_mod",
			},
			Prefix:  "ronni!ronni@ronni.tmi.twitch.tv",
			Command: "PRIVMSG",
			Params: []string{
				"#ronni",
				"Kappa Keepo Kappa",
			},
		},
		{
			Raw: `@badge-info=;badges=staff/1,bits/1000;bits=100;color=;display-name=ronni;emotes=;id=b34ccfc7-4977-403a-8a94-33c6bac34fb8;mod=0;room-id=1337;subscriber=0;tmi-sent-ts=1507246572675;turbo=1;user-id=1337;user-type=staff :ronni!ronni@ronni.tmi.twitch.tv PRIVMSG #ronni :cheer100`,
			Tags: map[string]string{
				"badge-info":   "",
				"badges":       "staff/1,bits/1000",
				"bits":         "100",
				"color":        "",
				"display-name": "ronni",
				"emotes":       "",
				"id":           "b34ccfc7-4977-403a-8a94-33c6bac34fb8",
				"mod":          "0",
				"room-id":      "1337",
				"subscriber":   "0",
				"tmi-sent-ts":  "1507246572675",
				"turbo":        "1",
				"user-id":      "1337",
				"user-type":    "staff",
			},
			Prefix:  "ronni!ronni@ronni.tmi.twitch.tv",
			Command: "PRIVMSG",
			Params: []string{
				"#ronni",
				"cheer100",
			},
		},
	}

	var parseIRCTests = make([]parseIRCTest, len(testIRCData))
	for i, v := range testIRCData {
		parseIRCTests[i] = parseIRCTest{v.Raw, v}
	}

	for i := range parseIRCTests {
		pIRCT := parseIRCTests[i]
		t.Run(pIRCT.in, func(t *testing.T) {
			t.Parallel()
			got, err := parseIRCMessage(pIRCT.in)
			if err != nil {
				t.Error(err)
			}
			want := pIRCT.out
			if !got.equals(want) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, want)
			}
		})
	}
}

func TestEscapeIRCTagValues(t *testing.T) {
	testTags := make(IRCTags)
	testTags["t1"] = `ronni\shas\ssubscribed\sfor\s6\smonths!`
	testTags["t2"] = `TWW2\sgifted\sa\sTier\s1\ssub\sto\sMr_Woodchuck!`
	testTags["t3"] = `An\sanonymous\suser\sgifted\sa\sTier\s1\ssub\sto\sTenureCalculator!\s`
	testTags["t4"] = `15\sraiders\sfrom\sTestChannel\shave\sjoined\n!`
	testTags["t5"] = `Seventoes\sis\snew\shere!`
	testTags["t6"] = `\\I\shave\n\sall\r\sthe\ssymbols\:\s`

	tests := []struct {
		key  string
		want string
	}{
		{"t1", "ronni has subscribed for 6 months!"},
		{"t2", "TWW2 gifted a Tier 1 sub to Mr_Woodchuck!"},
		{"t3", "An anonymous user gifted a Tier 1 sub to TenureCalculator!"},
		{"t4", "15 raiders from TestChannel have joined!"},
		{"t5", "Seventoes is new here!"},
		{"t6", `\I have all the symbols;`},
	}

	testTags.EscapeIRCTagValues()
	for _, v := range tests {
		if testTags[v.key] != v.want {
			t.Errorf("got: %v, want: %v\n", testTags[v.key], v.want)
		}
	}
}

func TestParseClearChatMessage(t *testing.T) {
	tests := []struct {
		in   string
		want ClearChatMessage
	}{
		{
			"@ban-duration=150;room-id=71092938;target-user-id=48215771;tmi-sent-ts=1568505600739 :tmi.twitch.tv CLEARCHAT #bobby :onche",
			ClearChatMessage{
				Channel:     "bobby",
				IRCType:     "CLEARCHAT",
				Text:        "onche timed out for 150 seconds in bobby",
				Type:        CLEARCHAT,
				BanDuration: time.Second * 150,
				Target:      "onche",
			},
		},
		{
			"@ban-duration=10;room-id=71092938;target-user-id=53211996;tmi-sent-ts=1568505608390 :tmi.twitch.tv CLEARCHAT #xqcow :haru_exc",
			ClearChatMessage{
				Channel:     "xqcow",
				IRCType:     "CLEARCHAT",
				Text:        "haru_exc timed out for 10 seconds in xqcow",
				Type:        CLEARCHAT,
				BanDuration: time.Second * 10,
				Target:      "haru_exc",
			},
		},
		{
			"@room-id=71092938;target-user-id=462385855;tmi-sent-ts=1568505916367 :tmi.twitch.tv CLEARCHAT #apocalypse :xmukkk",
			ClearChatMessage{
				Channel:     "apocalypse",
				IRCType:     "CLEARCHAT",
				Text:        "xmukkk was permanently banned in apocalypse",
				Type:        CLEARCHAT,
				BanDuration: -1,
				Target:      "xmukkk",
			},
		},
		{
			"@room-id=1234567;tmi-sent-ts=1234567 :tmi.twitch.tv CLEARCHAT #twitch",
			ClearChatMessage{
				Channel:     "twitch",
				IRCType:     "CLEARCHAT",
				Text:        "chat cleared in twitch",
				Type:        CLEARCHAT,
				BanDuration: -1,
				Target:      "",
			},
		},
	}

	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got := parseClearChatMessage(ircData)

		if got.Channel != test.want.Channel {
			t.Errorf("Channel: got %v, want %v", got.Channel, test.want.Channel)
		}
		if got.IRCType != test.want.IRCType {
			t.Errorf("IRCType: got %v, want %v", got.IRCType, test.want.IRCType)
		}
		if got.Text != test.want.Text {
			t.Errorf("Text: got %v, want %v", got.Text, test.want.Text)
		}
		if got.Type != test.want.Type {
			t.Errorf("Type: got %v, want %v", got.Type, test.want.Type)
		}
		if got.BanDuration != test.want.BanDuration {
			t.Errorf("BanDuration: got %v, want %v", got.BanDuration, test.want.BanDuration)
		}
		if got.Target != test.want.Target {
			t.Errorf("Target: got %v, want %v", got.Target, test.want.Target)
		}
	}
}
