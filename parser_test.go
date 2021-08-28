package tmi

import (
	"testing"
	"time"
)

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
				Channel:     "#bobby",
				IRCType:     "CLEARCHAT",
				Text:        "onche timed out for 150 seconds in #bobby",
				Type:        CLEARCHAT,
				BanDuration: time.Second * 150,
				Target:      "onche",
			},
		},
		{
			"@ban-duration=10;room-id=71092938;target-user-id=53211996;tmi-sent-ts=1568505608390 :tmi.twitch.tv CLEARCHAT #xqcow :haru_exc",
			ClearChatMessage{
				Channel:     "#xqcow",
				IRCType:     "CLEARCHAT",
				Text:        "haru_exc timed out for 10 seconds in #xqcow",
				Type:        CLEARCHAT,
				BanDuration: time.Second * 10,
				Target:      "haru_exc",
			},
		},
		{
			"@room-id=71092938;target-user-id=462385855;tmi-sent-ts=1568505916367 :tmi.twitch.tv CLEARCHAT #apocalypse :xmukkk",
			ClearChatMessage{
				Channel:     "#apocalypse",
				IRCType:     "CLEARCHAT",
				Text:        "xmukkk was permanently banned in #apocalypse",
				Type:        CLEARCHAT,
				BanDuration: -1,
				Target:      "xmukkk",
			},
		},
		{
			"@room-id=1234567;tmi-sent-ts=1234567 :tmi.twitch.tv CLEARCHAT #twitch",
			ClearChatMessage{
				Channel:     "#twitch",
				IRCType:     "CLEARCHAT",
				Text:        "chat cleared in #twitch",
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

		assertStringsEqual(t, "Channel", got.Channel, test.want.Channel)
		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertStringsEqual(t, "Text", got.Text, test.want.Text)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		assertDurationsEqual(t, "BanDuration", got.BanDuration, test.want.BanDuration)
		assertStringsEqual(t, "Target", got.Target, test.want.Target)
	}
}

func TestParseClearMsgMessage(t *testing.T) {
	tests := []struct {
		in   string
		want ClearMsgMessage
	}{
		{
			"@login=ronni;target-msg-id=abc-123-def :tmi.twitch.tv CLEARMSG #dallas :HeyGuys",
			ClearMsgMessage{
				Channel:     "#dallas",
				IRCType:     "CLEARMSG",
				Text:        "HeyGuys",
				Type:        CLEARMSG,
				Login:       "ronni",
				TargetMsgID: "abc-123-def",
			},
		},
		{
			"@login=<login>;target-msg-id=<target-msg-id> :tmi.twitch.tv CLEARMSG #<channel> :<message>",
			ClearMsgMessage{
				Channel:     "#<channel>",
				IRCType:     "CLEARMSG",
				Text:        "<message>",
				Type:        CLEARMSG,
				Login:       "<login>",
				TargetMsgID: "<target-msg-id>",
			},
		},
	}

	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got := parseClearMsgMessage(ircData)

		assertStringsEqual(t, "Channel", got.Channel, test.want.Channel)
		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertStringsEqual(t, "Text", got.Text, test.want.Text)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		assertStringsEqual(t, "Login", got.Login, test.want.Login)
		assertStringsEqual(t, "TargetMsgID", got.TargetMsgID, test.want.TargetMsgID)
	}
}

func TestParseGlobalUserstateMessage(t *testing.T) {
	tests := []struct {
		in   string
		want GlobalUserstateMessage
	}{
		{
			"@badge-info=<badge-info>;badges=<badge>/1;color=<color>;display-name=<display-name>;emote-sets=<emote-sets>;turbo=<turbo>;user-id=<user-id>;user-type=<user-type> :tmi.twitch.tv GLOBALUSERSTATE",
			GlobalUserstateMessage{
				IRCType:   "GLOBALUSERSTATE",
				Type:      GLOBALUSERSTATE,
				EmoteSets: []string{"<emote-sets>"},
				User: &User{
					BadgeInfo: "<badge-info>",
					Badges: []Badge{
						{"<badge>", 1},
					},
					Broadcaster: false,
					Color:       "<color>",
					DisplayName: "<display-name>",
					Mod:         false,
					Name:        "<display-name>",
					Subscriber:  false,
					Turbo:       false,
					ID:          "<user-id>",
					UserType:    "<user-type>",
					VIP:         false,
				},
			},
		},
		{
			"@badge-info=subscriber/8;badges=subscriber/6;color=#0D4200;display-name=dallas;emote-sets=0,33,50,237,793,2126,3517,4578,5569,9400,10337,12239;turbo=0;user-id=1337;user-type=admin :tmi.twitch.tv GLOBALUSERSTATE",
			GlobalUserstateMessage{
				IRCType:   "GLOBALUSERSTATE",
				Type:      GLOBALUSERSTATE,
				EmoteSets: []string{"0", "33", "50", "237", "793", "2126", "3517", "4578", "5569", "9400", "10337", "12239"},
				User: &User{
					BadgeInfo: "subscriber/8",
					Badges: []Badge{
						{"subscriber", 6},
					},
					Broadcaster: false,
					Color:       "#0D4200",
					DisplayName: "dallas",
					Mod:         false,
					Name:        "dallas",
					Subscriber:  true,
					Turbo:       false,
					ID:          "1337",
					UserType:    "admin",
					VIP:         false,
				},
			},
		},
	}
	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got := parseGlobalUserstateMessage(ircData)

		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		assertStringSlicesEqual(t, "EmoteSets", got.EmoteSets, test.want.EmoteSets)
		if !got.User.equals(test.want.User) {
			t.Errorf("User: got %v, want %v", got.User, test.want.User)
		}
	}
}

func TestParseHostTargetMessage(t *testing.T) {
	tests := []struct {
		in   string
		want HostTargetMessage
	}{
		{
			":tmi.twitch.tv HOSTTARGET #hosting_channel :-",
			HostTargetMessage{
				Channel: "#hosting_channel",
				IRCType: "HOSTTARGET",
				Text:    "#hosting_channel exited host mode",
				Type:    HOSTTARGET,
				Hosted:  "",
				Viewers: 0,
			},
		},
		{
			":tmi.twitch.tv HOSTTARGET #hosting_channel :- 5",
			HostTargetMessage{
				Channel: "#hosting_channel",
				IRCType: "HOSTTARGET",
				Text:    "#hosting_channel exited host mode with 5 viewers",
				Type:    HOSTTARGET,
				Hosted:  "",
				Viewers: 5,
			},
		},
		{
			":tmi.twitch.tv HOSTTARGET #hosting_channel :channel",
			HostTargetMessage{
				Channel: "#hosting_channel",
				IRCType: "HOSTTARGET",
				Text:    "#hosting_channel is now hosting channel",
				Type:    HOSTTARGET,
				Hosted:  "channel",
				Viewers: 0,
			},
		},
		{
			":tmi.twitch.tv HOSTTARGET #hosting_channel :channel 16",
			HostTargetMessage{
				Channel: "#hosting_channel",
				IRCType: "HOSTTARGET",
				Text:    "#hosting_channel is now hosting channel with 16 viewers",
				Type:    HOSTTARGET,
				Hosted:  "channel",
				Viewers: 16,
			},
		},
	}

	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got := parseHostTargetMessage(ircData)

		assertStringsEqual(t, "Channel", got.Channel, test.want.Channel)
		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertStringsEqual(t, "Text", got.Text, test.want.Text)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		assertStringsEqual(t, "Hosted", got.Hosted, test.want.Hosted)
		assertIntsEqual(t, "Viewers", got.Viewers, test.want.Viewers)
	}
}

func TestParseNoticeMessage(t *testing.T) {
	tests := []struct {
		in   string
		want NoticeMessage
	}{
		{
			"@msg-id=<msg-id> :tmi.twitch.tv NOTICE #<channel> :<message>",
			NoticeMessage{
				Channel: "#<channel>",
				IRCType: "NOTICE",
				Text:    "<message>",
				Type:    NOTICE,
				Enabled: false,
				Mods:    []string{},
				MsgID:   "<msg-id>",
				Notice:  "notice",
				VIPs:    []string{},
			},
		},
		{
			"@msg-id=slow_off :tmi.twitch.tv NOTICE #dallas :This room is no longer in slow mode.",
			NoticeMessage{
				Channel: "#dallas",
				IRCType: "NOTICE",
				Text:    "This room is no longer in slow mode.",
				Type:    NOTICE,
				Enabled: false,
				Mods:    []string{},
				MsgID:   "slow_off",
				Notice:  "notice",
				VIPs:    []string{},
			},
		},
		{
			"@msg-id=r9k_on :tmi.twitch.tv NOTICE #achannel :This room is now in unique-chat mode.",
			NoticeMessage{
				Channel: "#achannel",
				IRCType: "NOTICE",
				Text:    "This room is now in unique-chat mode.",
				Type:    NOTICE,
				Enabled: true,
				Mods:    []string{},
				MsgID:   "r9k_on",
				Notice:  "uniquechat",
				VIPs:    []string{},
			},
		},
		{
			"@msg-id=r9k_off :tmi.twitch.tv NOTICE #somechannel :This room is no longer in unique-chat mode.",
			NoticeMessage{
				Channel: "#somechannel",
				IRCType: "NOTICE",
				Text:    "This room is no longer in unique-chat mode.",
				Type:    NOTICE,
				Enabled: false,
				Mods:    []string{},
				MsgID:   "r9k_off",
				Notice:  "uniquechat",
				VIPs:    []string{},
			},
		},
		{
			"@msg-id=emote_only_on :tmi.twitch.tv NOTICE #ch :This room is now in emote-only mode.",
			NoticeMessage{
				Channel: "#ch",
				IRCType: "NOTICE",
				Text:    "This room is now in emote-only mode.",
				Type:    NOTICE,
				Enabled: true,
				Mods:    []string{},
				MsgID:   "emote_only_on",
				Notice:  "emoteonly",
				VIPs:    []string{},
			},
		},
		{
			"@msg-id=emote_only_off :tmi.twitch.tv NOTICE #itsachannel :This room is no longer in emote-only mode.",
			NoticeMessage{
				Channel: "#itsachannel",
				IRCType: "NOTICE",
				Text:    "This room is no longer in emote-only mode.",
				Type:    NOTICE,
				Enabled: false,
				Mods:    []string{},
				MsgID:   "emote_only_off",
				Notice:  "emoteonly",
				VIPs:    []string{},
			},
		},
		{
			"@msg-id=subs_on :tmi.twitch.tv NOTICE #yep :This room is now in subscribers-only mode.",
			NoticeMessage{
				Channel: "#yep",
				IRCType: "NOTICE",
				Text:    "This room is now in subscribers-only mode.",
				Type:    NOTICE,
				Enabled: true,
				Mods:    []string{},
				MsgID:   "subs_on",
				Notice:  "subonly",
				VIPs:    []string{},
			},
		},
		{
			"@msg-id=subs_off :tmi.twitch.tv NOTICE #nope :This room is no longer in subscribers-only mode.",
			NoticeMessage{
				Channel: "#nope",
				IRCType: "NOTICE",
				Text:    "This room is no longer in subscribers-only mode.",
				Type:    NOTICE,
				Enabled: false,
				Mods:    []string{},
				MsgID:   "subs_off",
				Notice:  "subonly",
				VIPs:    []string{},
			},
		},
	}

	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got, err := parseNoticeMessage(ircData)
		if err != nil {
			t.Error(err)
		}

		assertStringsEqual(t, "Channel", got.Channel, test.want.Channel)
		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertStringsEqual(t, "Text", got.Text, test.want.Text)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		assertBoolsEqual(t, "Enabled", got.Enabled, test.want.Enabled)
		assertStringSlicesEqual(t, "Mods", got.Mods, test.want.Mods)
		assertStringsEqual(t, "MsgID", got.MsgID, test.want.MsgID)
		assertStringsEqual(t, "Notice", got.Notice, test.want.Notice)
		assertStringSlicesEqual(t, "VIPs", got.VIPs, test.want.VIPs)
	}
}

func assertStringsEqual(t *testing.T, name, s1, s2 string) {
	if s1 != s2 {
		t.Errorf("%v: got %v, want %v", name, s1, s2)
	}
}

func assertStringSlicesEqual(t *testing.T, name string, s1, s2 []string) {
	if len(s1) != len(s2) {
		t.Errorf("%v: len(got) %v, len(want) %v", name, len(s1), len(s2))
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			t.Errorf("%v[%v]: got %v, want %v", name, i, s1[i], s2[i])
		}
	}
}

func assertMessageTypesEqual(t *testing.T, t1, t2 MessageType) {
	if t1 != t2 {
		t.Errorf("Type: got %v, want %v", t1, t2)
	}
}

func assertDurationsEqual(t *testing.T, name string, d1, d2 time.Duration) {
	if d1 != d2 {
		t.Errorf("%v: got %v, want %v", name, d1, d2)
	}
}

func assertIntsEqual(t *testing.T, name string, i1, i2 int) {
	if i1 != i2 {
		t.Errorf("%v: got %v, want %v", name, i1, i2)
	}
}

func assertBoolsEqual(t *testing.T, name string, b1, b2 bool) {
	if b1 != b2 {
		t.Errorf("%v: got %v, want %v", name, b1, b2)
	}
}

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

func (u1 *User) equals(u2 *User) bool {
	if u1.BadgeInfo != u2.BadgeInfo {
		return false
	}
	for i, badge := range u1.Badges {
		if badge != u2.Badges[i] {
			return false
		}
	}
	if u1.Broadcaster != u2.Broadcaster {
		return false
	}
	if u1.Color != u2.Color {
		return false
	}
	if u1.DisplayName != u2.DisplayName {
		return false
	}
	if u1.Mod != u2.Mod {
		return false
	}
	if u1.Name != u2.Name {
		return false
	}
	if u1.Subscriber != u2.Subscriber {
		return false
	}
	if u1.Turbo != u2.Turbo {
		return false
	}
	if u1.ID != u2.ID {
		return false
	}
	if u1.UserType != u2.UserType {
		return false
	}
	if u1.VIP != u2.VIP {
		return false
	}
	return true
}
