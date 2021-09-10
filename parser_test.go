package tmi

import (
	"fmt"
	"testing"
	"time"
)

func TestParseIRCMessage(t *testing.T) {
	t.Parallel()
	type parseIRCTest struct {
		in  string
		out IRCData
	}
	var testIRCData = []IRCData{
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
			assertIRCDataEqual(t, &got, &want)
		})
	}
}

func TestEscapeIRCTagValues(t *testing.T) {
	tests := make(IRCTags)
	tests["t1"] = `ronni\shas\ssubscribed\sfor\s6\smonths!`
	tests["t2"] = `TWW2\sgifted\sa\sTier\s1\ssub\sto\sMr_Woodchuck!`
	tests["t3"] = `An\sanonymous\suser\sgifted\sa\sTier\s1\ssub\sto\sTenureCalculator!\s`
	tests["t4"] = `15\sraiders\sfrom\sTestChannel\shave\sjoined\n!`
	tests["t5"] = `Seventoes\sis\snew\shere!`
	tests["t6"] = `\\I\shave\sall\r\n\sthe\ssymbols\:\s`

	want := IRCTags{
		"t1": "ronni has subscribed for 6 months!",
		"t2": "TWW2 gifted a Tier 1 sub to Mr_Woodchuck!",
		"t3": "An anonymous user gifted a Tier 1 sub to TenureCalculator!",
		"t4": "15 raiders from TestChannel have joined\n!",
		"t5": "Seventoes is new here!",
		"t6": "\\I have all\r\n the symbols;",
	}

	tests.EscapeIRCTagValues()
	assertStringMapsEqual(t, "Tags", tests, want)
}

func TestParseTimeStamp(t *testing.T) {
	tests := []struct {
		in   string
		want time.Time
	}{
		{"1568505600739", time.Date(2019, time.September, 14, 20, 0, 0, 739*1e6, time.Local)},
		{"1568505608390", time.Date(2019, time.September, 14, 20, 0, 8, 390*1e6, time.Local)},
		{"1630887934441", time.Date(2021, time.September, 5, 20, 25, 34, 441*1e6, time.Local)},
	}

	for _, test := range tests {
		got := ParseTimeStamp(test.in)
		want := test.want
		if got != want {
			t.Errorf("ParseTimeStamp: got %v, want %v", got, want)
		}
	}
}

func TestParseReplyParentMessage(t *testing.T) {
	tests := []struct {
		in   IRCTags
		want ReplyParentMsg
	}{
		{
			IRCTags{
				"reply-parent-msg-id": "b34ccfc7-4977-403a-8a94-33c6bac34fb8",
			},
			ReplyParentMsg{ID: "b34ccfc7-4977-403a-8a94-33c6bac34fb8"},
		},
		{
			IRCTags{
				"reply-parent-display-name": "ThisIsSparta",
				"reply-parent-msg-id":       "987654321",
				"reply-parent-msg-body":     "pigs are able to fly now",
				"reply-parent-user-id":      "123456789",
				"reply-parent-user-login":   "thisissparta",
			},
			ReplyParentMsg{
				"ThisIsSparta",
				"987654321",
				"pigs are able to fly now",
				"123456789",
				"thisissparta",
			},
		},
	}

	for _, test := range tests {
		got := ParseReplyParentMessage(test.in)
		want := test.want
		if got != want {
			t.Errorf("ParseReplyParentMessage: got %v, want %v", got, want)
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
		assertUsersEqual(t, got.User, test.want.User)
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

func TestParseRoomstateMessage(t *testing.T) {
	tests := []struct {
		in   string
		want RoomstateMessage
	}{
		{
			"@emote-only=<emote-only>;followers-only=<followers-only>;r9k=<r9k>;rituals=<ritual>;slow=<slow>;subs-only=<subs-only> :tmi.twitch.tv ROOMSTATE #<channel>",
			RoomstateMessage{
				Channel: "#<channel>",
				IRCType: "ROOMSTATE",
				Type:    ROOMSTATE,
				States: map[string]RoomState{
					"emote-only":     {},
					"followers-only": {},
					"r9k":            {},
					"rituals":        {},
					"slow":           {},
					"subs-only":      {},
				},
			},
		},
		{
			"@emote-only=0;followers-only=0;r9k=0;slow=0;subs-only=0 :tmi.twitch.tv ROOMSTATE #dallas",
			RoomstateMessage{
				Channel: "#dallas",
				IRCType: "ROOMSTATE",
				Type:    ROOMSTATE,
				States: map[string]RoomState{
					"emote-only":     {false, 0},
					"followers-only": {true, 0},
					"r9k":            {false, 0},
					"slow":           {false, 0},
					"subs-only":      {false, 0},
				},
			},
		},
		{
			"@slow=10 :tmi.twitch.tv ROOMSTATE #dallas",
			RoomstateMessage{
				Channel: "#dallas",
				IRCType: "ROOMSTATE",
				Type:    ROOMSTATE,
				States: map[string]RoomState{
					"slow": {true, time.Second * 10},
				},
			},
		},
		{
			"@followers-only=-1 :tmi.twitch.tv ROOMSTATE #whoever",
			RoomstateMessage{
				Channel: "#whoever",
				IRCType: "ROOMSTATE",
				Type:    ROOMSTATE,
				States: map[string]RoomState{
					"followers-only": {false, 0},
				},
			},
		},
		{
			"@followers-only=1 :tmi.twitch.tv ROOMSTATE #anyone",
			RoomstateMessage{
				Channel: "#anyone",
				IRCType: "ROOMSTATE",
				Type:    ROOMSTATE,
				States: map[string]RoomState{
					"followers-only": {true, time.Minute},
				},
			},
		},
	}

	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got := parseRoomstateMessage(ircData)

		assertStringsEqual(t, "Channel", got.Channel, test.want.Channel)
		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		if len(got.States) != len(test.want.States) {
			t.Errorf("States: len(got) %v, len(want) %v", len(got.States), len(test.want.States))
		}
		for k, wv := range test.want.States {
			if gv, ok := got.States[k]; ok {
				if gv != wv {
					t.Errorf("States[%v]: got %v, want %v", k, gv, wv)
				}
			} else {
				t.Errorf("States[%v] not created in got, want %v", k, wv)
			}
		}
	}
}

func TestParseUsernoticeMessage(t *testing.T) {
	tests := []struct {
		in   string
		want UsernoticeMessage
	}{
		{
			`@badge-info=<badge-info>;badges=<badges>;color=<color>;display-name=<display-name>;emotes=<emotes>;id=<id-of-msg>;login=<user>;mod=<mod>;msg-id=<msg-id>;room-id=<room-id>;subscriber=<subscriber>;system-msg=<system-msg>;tmi-sent-ts=<timestamp>;turbo=<turbo>;user-id=<user-id>;user-type=<user-type> :tmi.twitch.tv USERNOTICE #<channel> :<message>`,
			UsernoticeMessage{
				Channel:   "#<channel>",
				IRCType:   "USERNOTICE",
				Text:      "<message>",
				Type:      USERNOTICE,
				Emotes:    []Emote{},
				ID:        "<id-of-msg>",
				MsgID:     "<msg-id>",
				MsgParams: IRCTags{},
				SystemMsg: `<system-msg>`,
				User: &User{
					BadgeInfo:   "<badge-info>",
					Badges:      []Badge{},
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
			`@badge-info=;badges=staff/1,broadcaster/1,turbo/1;color=#008000;display-name=ronni;emotes=;id=db25007f-7a18-43eb-9379-80131e44d633;login=ronni;mod=0;msg-id=resub;msg-param-cumulative-months=6;msg-param-streak-months=2;msg-param-should-share-streak=1;msg-param-sub-plan=Prime;msg-param-sub-plan-name=Prime;room-id=1337;subscriber=1;system-msg=ronni\shas\ssubscribed\sfor\s6\smonths!;tmi-sent-ts=1507246572675;turbo=1;user-id=1337;user-type=staff :tmi.twitch.tv USERNOTICE #dallas :Great stream -- keep it up!`,
			UsernoticeMessage{
				Channel: "#dallas",
				IRCType: "USERNOTICE",
				Text:    "Great stream -- keep it up!",
				Type:    USERNOTICE,
				Emotes:  []Emote{},
				ID:      "db25007f-7a18-43eb-9379-80131e44d633",
				MsgID:   "resub",
				MsgParams: IRCTags{
					"msg-param-cumulative-months":   `6`,
					"msg-param-streak-months":       `2`,
					"msg-param-should-share-streak": `1`,
					"msg-param-sub-plan":            `Prime`,
					"msg-param-sub-plan-name":       `Prime`,
				},
				SystemMsg: `ronni\shas\ssubscribed\sfor\s6\smonths!`,
				User: &User{
					BadgeInfo: "",
					Badges: []Badge{
						{"staff", 1},
						{"broadcaster", 1},
						{"turbo", 1},
					},
					Broadcaster: true,
					Color:       "#008000",
					DisplayName: "ronni",
					Mod:         false,
					Name:        "ronni",
					Subscriber:  true,
					Turbo:       true,
					ID:          "1337",
					UserType:    "staff",
					VIP:         false,
				},
			},
		},
		{
			`@badge-info=;badges=staff/1,premium/1;color=#0000FF;display-name=TWW2;emotes=;id=e9176cd8-5e22-4684-ad40-ce53c2561c5e;login=tww2;mod=0;msg-id=subgift;msg-param-months=1;msg-param-recipient-display-name=Mr_Woodchuck;msg-param-recipient-id=89614178;msg-param-recipient-name=mr_woodchuck;msg-param-sub-plan-name=House\sof\sNyoro~n;msg-param-sub-plan=1000;room-id=19571752;subscriber=0;system-msg=TWW2\sgifted\sa\sTier\s1\ssub\sto\sMr_Woodchuck!;tmi-sent-ts=1521159445153;turbo=0;user-id=13405587;user-type=staff :tmi.twitch.tv USERNOTICE #forstycup`,
			UsernoticeMessage{
				Channel: "#forstycup",
				IRCType: "USERNOTICE",
				Text:    "",
				Type:    USERNOTICE,
				Emotes:  []Emote{},
				ID:      "e9176cd8-5e22-4684-ad40-ce53c2561c5e",
				MsgID:   "subgift",
				MsgParams: IRCTags{
					"msg-param-months":                 `1`,
					"msg-param-recipient-display-name": `Mr_Woodchuck`,
					"msg-param-recipient-id":           `89614178`,
					"msg-param-recipient-name":         `mr_woodchuck`,
					"msg-param-sub-plan-name":          `House\sof\sNyoro~n`,
					"msg-param-sub-plan":               `1000`,
				},
				SystemMsg: `TWW2\sgifted\sa\sTier\s1\ssub\sto\sMr_Woodchuck!`,
				User: &User{
					BadgeInfo: "",
					Badges: []Badge{
						{"staff", 1},
						{"premium", 1},
					},
					Broadcaster: false,
					Color:       "#0000FF",
					DisplayName: "TWW2",
					Mod:         false,
					Name:        "tww2",
					Subscriber:  false,
					Turbo:       false,
					ID:          "13405587",
					UserType:    "staff",
					VIP:         false,
				},
			},
		},
		{
			`@badge-info=;badges=broadcaster/1,subscriber/6;color=;display-name=qa_subs_partner;emotes=;flags=;id=b1818e3c-0005-490f-ad0a-804957ddd760;login=qa_subs_partner;mod=0;msg-id=anonsubgift;msg-param-months=3;msg-param-recipient-display-name=TenureCalculator;msg-param-recipient-id=135054130;msg-param-recipient-user-name=tenurecalculator;msg-param-sub-plan-name=t111;msg-param-sub-plan=1000;room-id=196450059;subscriber=1;system-msg=An\sanonymous\suser\sgifted\sa\sTier\s1\ssub\sto\sTenureCalculator!\s;tmi-sent-ts=1542063432068;turbo=0;user-id=196450059;user-type= :tmi.twitch.tv USERNOTICE #qa_subs_partner`,
			UsernoticeMessage{
				Channel: "#qa_subs_partner",
				IRCType: "USERNOTICE",
				Text:    "",
				Type:    USERNOTICE,
				Emotes:  []Emote{},
				ID:      "b1818e3c-0005-490f-ad0a-804957ddd760",
				MsgID:   "anonsubgift",
				MsgParams: IRCTags{
					"msg-param-months":                 `3`,
					"msg-param-recipient-display-name": `TenureCalculator`,
					"msg-param-recipient-id":           `135054130`,
					"msg-param-recipient-user-name":    `tenurecalculator`,
					"msg-param-sub-plan-name":          `t111`,
					"msg-param-sub-plan":               `1000`,
				},
				SystemMsg: `An\sanonymous\suser\sgifted\sa\sTier\s1\ssub\sto\sTenureCalculator!\s`,
				User: &User{
					BadgeInfo: "",
					Badges: []Badge{
						{"broadcaster", 1},
						{"subscriber", 6},
					},
					Broadcaster: true,
					Color:       "",
					DisplayName: "qa_subs_partner",
					Mod:         false,
					Name:        "qa_subs_partner",
					Subscriber:  true,
					Turbo:       false,
					ID:          "196450059",
					UserType:    "",
					VIP:         false,
				},
			},
		},
		{
			`@badge-info=;badges=turbo/1;color=#9ACD32;display-name=TestChannel;emotes=;id=3d830f12-795c-447d-af3c-ea05e40fbddb;login=testchannel;mod=0;msg-id=raid;msg-param-displayName=TestChannel;msg-param-login=testchannel;msg-param-viewerCount=15;room-id=56379257;subscriber=0;system-msg=15\sraiders\sfrom\sTestChannel\shave\sjoined\n!;tmi-sent-ts=1507246572675;tmi-sent-ts=1507246572675;turbo=1;user-id=123456;user-type= :tmi.twitch.tv USERNOTICE #othertestchannel`,
			UsernoticeMessage{
				Channel: "#othertestchannel",
				IRCType: "USERNOTICE",
				Text:    "",
				Type:    USERNOTICE,
				Emotes:  []Emote{},
				ID:      "3d830f12-795c-447d-af3c-ea05e40fbddb",
				MsgID:   "raid",
				MsgParams: IRCTags{
					"msg-param-displayName": `TestChannel`,
					"msg-param-login":       `testchannel`,
					"msg-param-viewerCount": `15`,
				},
				SystemMsg: `15\sraiders\sfrom\sTestChannel\shave\sjoined\n!`,
				User: &User{
					BadgeInfo: "",
					Badges: []Badge{
						{"turbo", 1},
					},
					Broadcaster: false,
					Color:       "#9ACD32",
					DisplayName: "TestChannel",
					Mod:         false,
					Name:        "testchannel",
					Subscriber:  false,
					Turbo:       true,
					ID:          "123456",
					UserType:    "",
					VIP:         false,
				},
			},
		},
		{
			`@badge-info=;badges=;color=;display-name=SevenTest1;emotes=30259:0-6;id=37feed0f-b9c7-4c3a-b475-21c6c6d21c3d;login=seventest1;mod=0;msg-id=ritual;msg-param-ritual-name=new_chatter;room-id=6316121;subscriber=0;system-msg=Seventoes\sis\snew\shere!;tmi-sent-ts=1508363903826;turbo=0;user-id=131260580;user-type= :tmi.twitch.tv USERNOTICE #seventoes :HeyGuys`,
			UsernoticeMessage{
				Channel: "#seventoes",
				IRCType: "USERNOTICE",
				Text:    "HeyGuys",
				Type:    USERNOTICE,
				Emotes:  []Emote{{"30259", "HeyGuys", []EmotePosition{{0, 6}}}},
				ID:      "37feed0f-b9c7-4c3a-b475-21c6c6d21c3d",
				MsgID:   "ritual",
				MsgParams: IRCTags{
					"msg-param-ritual-name": `new_chatter`,
				},
				SystemMsg: `Seventoes\sis\snew\shere!`,
				User: &User{
					BadgeInfo:   "",
					Badges:      []Badge{},
					Broadcaster: false,
					Color:       "",
					DisplayName: "SevenTest1",
					Mod:         false,
					Name:        "seventest1",
					Subscriber:  false,
					Turbo:       false,
					ID:          "131260580",
					UserType:    "",
					VIP:         false,
				},
			},
		},
	}

	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got := parseUsernoticeMessage(ircData)

		assertStringsEqual(t, "Channel", got.Channel, test.want.Channel)
		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertStringsEqual(t, "Text", got.Text, test.want.Text)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		assertEmoteSlicesEqual(t, got.Emotes, test.want.Emotes)
		assertStringMapsEqual(t, "MsgParams", got.MsgParams, test.want.MsgParams)
		assertStringsEqual(t, "SystemMsg", got.SystemMsg, test.want.SystemMsg)
		assertUsersEqual(t, got.User, test.want.User)
	}
}

func TestParseUserstateMessage(t *testing.T) {
	tests := []struct {
		in   string
		want UserstateMessage
	}{
		{
			"@badge-info=<badge-info>;badges=<badges>;color=<color>;display-name=<display-name>;emote-sets=<emote-sets>;mod=<mod>;subscriber=<subscriber>;turbo=<turbo>;user-type=<user-type> :tmi.twitch.tv USERSTATE #<channel>",
			UserstateMessage{
				Channel:   "#<channel>",
				IRCType:   "USERSTATE",
				Type:      USERSTATE,
				EmoteSets: []string{"<emote-sets>"},
				User: &User{
					BadgeInfo:   "<badge-info>",
					Badges:      []Badge{},
					Broadcaster: false,
					Color:       "<color>",
					DisplayName: "<display-name>",
					Mod:         false,
					Name:        "<display-name>",
					Subscriber:  false,
					Turbo:       false,
					ID:          "",
					UserType:    "<user-type>",
					VIP:         false,
				},
			},
		},
		{
			"@badge-info=;badges=staff/1;color=#0D4200;display-name=ronni;emote-sets=0,33,50,237,793,2126,3517,4578,5569,9400,10337,12239;mod=1;subscriber=1;turbo=1;user-type=staff :tmi.twitch.tv USERSTATE #dallas",
			UserstateMessage{
				Channel:   "#dallas",
				IRCType:   "USERSTATE",
				Type:      USERSTATE,
				EmoteSets: []string{"0", "33", "50", "237", "793", "2126", "3517", "4578", "5569", "9400", "10337", "12239"},
				User: &User{
					BadgeInfo: "",
					Badges: []Badge{
						{"staff", 1},
					},
					Broadcaster: false,
					Color:       "#0D4200",
					DisplayName: "ronni",
					Mod:         true,
					Name:        "ronni",
					Subscriber:  true,
					Turbo:       true,
					ID:          "",
					UserType:    "staff",
					VIP:         false,
				},
			},
		},
		{
			"@badge-info=;badges=moderator/1;color=#00FF7F;display-name=testerester;emote-sets=0,564265402,592920959,610186276;mod=1;subscriber=0;user-type=mod :tmi.twitch.tv USERSTATE #testing",
			UserstateMessage{
				Channel:   "#testing",
				IRCType:   "USERSTATE",
				Type:      USERSTATE,
				EmoteSets: []string{"0", "564265402", "592920959", "610186276"},
				User: &User{
					BadgeInfo: "",
					Badges: []Badge{
						{"moderator", 1},
					},
					Broadcaster: false,
					Color:       "#00FF7F",
					DisplayName: "testerester",
					Mod:         true,
					Name:        "testerester",
					Subscriber:  false,
					Turbo:       false,
					ID:          "",
					UserType:    "mod",
					VIP:         false,
				},
			},
		},
	}

	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got := parseUserstateMessage(ircData)

		assertStringsEqual(t, "Channel", got.Channel, test.want.Channel)
		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		assertStringSlicesEqual(t, "EmoteSets", got.EmoteSets, test.want.EmoteSets)
		assertUsersEqual(t, got.User, test.want.User)
	}
}

func TestParseNamesMessage(t *testing.T) {
	tests := []struct {
		in   string
		want NamesMessage
	}{
		{
			":<client>.tmi.twitch.tv 353 client = #<channel> :<client>",
			NamesMessage{
				Channel: "#<channel>",
				IRCType: "353",
				Type:    NAMES,
				Users:   []string{"<client>"},
			},
		},
		{
			":you.tmi.twitch.tv 353 client = #testchannel :you them1 them2 them3",
			NamesMessage{
				Channel: "#testchannel",
				IRCType: "353",
				Type:    NAMES,
				Users:   []string{"you", "them1", "them2", "them3"},
			},
		},
	}

	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got := parseNamesMessage(ircData)

		assertStringsEqual(t, "Channel", got.Channel, test.want.Channel)
		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		assertStringSlicesEqual(t, "Users", got.Users, test.want.Users)
	}
}

func TestParseJoinMessage(t *testing.T) {
	tests := []struct {
		in   string
		want JoinMessage
	}{
		{
			":<username>!<username>@<username>.tmi.twitch.tv JOIN #<channel>",
			JoinMessage{
				Channel:  "#<channel>",
				IRCType:  "JOIN",
				Type:     JOIN,
				Username: "<username>",
			},
		},
	}

	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got := parseJoinMessage(ircData)

		assertStringsEqual(t, "Channel", got.Channel, test.want.Channel)
		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		assertStringsEqual(t, "Username", got.Username, test.want.Username)
	}
}

func TestParsePartMessage(t *testing.T) {
	tests := []struct {
		in   string
		want PartMessage
	}{
		{
			":<username>!<username>@<username>.tmi.twitch.tv PART #<channel>",
			PartMessage{
				Channel:  "#<channel>",
				IRCType:  "PART",
				Type:     PART,
				Username: "<username>",
			},
		},
	}

	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got := parsePartMessage(ircData)

		assertStringsEqual(t, "Channel", got.Channel, test.want.Channel)
		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		assertStringsEqual(t, "Username", got.Username, test.want.Username)
	}
}

func TestParsePingMessage(t *testing.T) {
	tests := []struct {
		in   string
		want PingMessage
	}{
		{
			"PING",
			PingMessage{
				IRCType: "PING",
				Type:    PING,
				Text:    "",
			},
		},
		{
			"PING :tmi.twitch.tv",
			PingMessage{
				IRCType: "PING",
				Type:    PING,
				Text:    "tmi.twitch.tv",
			},
		},
	}

	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got := parsePingMessage(ircData)

		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		assertStringsEqual(t, "Text", got.Text, test.want.Text)
	}
}

func TestParsePongMessage(t *testing.T) {
	tests := []struct {
		in   string
		want PongMessage
	}{
		{
			"PONG",
			PongMessage{
				IRCType: "PONG",
				Type:    PONG,
				Text:    "",
			},
		},
		{
			":tmi.twitch.tv PONG tmi.twitch.tv :" + pingSignature,
			PongMessage{
				IRCType: "PONG",
				Type:    PONG,
				Text:    pingSignature,
			},
		},
	}

	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got := parsePongMessage(ircData)

		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		assertStringsEqual(t, "Text", got.Text, test.want.Text)
	}
}

func TestParsePrivateMessage(t *testing.T) {
	tests := []struct {
		in   string
		want PrivateMessage
	}{
		{
			"@badge-info=;badges=premium/1;client-nonce=b837cca5074aaa5eb4482e5df50b3e6a;color=;display-name=Banjana;emotes=;flags=;id=23c201b3-94e9-4d28-8f71-baac329af81c;mod=0;room-id=132230344;subscriber=0;tmi-sent-ts=1630888435197;turbo=0;user-id=138657205;user-type= :banjana!banjana@banjana.tmi.twitch.tv PRIVMSG #moistcr1tikal :Xrd is insane",
			PrivateMessage{
				Channel: "#moistcr1tikal",
				IRCType: "PRIVMSG",
				Type:    PRIVMSG,
				Text:    "Xrd is insane",
				Action:  false,
				Bits:    0,
				Emotes:  []Emote{},
				ID:      "23c201b3-94e9-4d28-8f71-baac329af81c",
				Reply:   false,
				User: &User{
					BadgeInfo: "",
					Badges: []Badge{
						{"premium", 1},
					},
					Broadcaster: false,
					Color:       "",
					DisplayName: "Banjana",
					Mod:         false,
					Name:        "banjana",
					Subscriber:  false,
					Turbo:       false,
					ID:          "138657205",
					UserType:    "",
					VIP:         false,
				},
			},
		},
		{
			"@badge-info=subscriber/4;badges=subscriber/3,glitchcon2020/1;client-nonce=0c57e7357cbea005b349f24ed0bdbf15;color=#0000FF;display-name=Ridz_;emote-only=1;emotes=303446392:0-3;flags=;id=9d8fca2e-2924-4b50-9655-2e0921d73eb9;mod=0;room-id=207813352;subscriber=1;tmi-sent-ts=1630888038100;turbo=0;user-id=138035491;user-type= :ridz_!ridz_@ridz_.tmi.twitch.tv PRIVMSG #hasanabi :hasO",
			PrivateMessage{
				Channel: "#hasanabi",
				IRCType: "PRIVMSG",
				Type:    PRIVMSG,
				Text:    "hasO",
				Action:  false,
				Bits:    0,
				Emotes: []Emote{
					{"303446392", "hasO", []EmotePosition{{0, 3}}},
				},
				ID:    "9d8fca2e-2924-4b50-9655-2e0921d73eb9",
				Reply: false,
				User: &User{
					BadgeInfo: "subscriber/4",
					Badges: []Badge{
						{"subscriber", 3},
						{"glitchcon2020", 1},
					},
					Broadcaster: false,
					Color:       "#0000FF",
					DisplayName: "Ridz_",
					Mod:         false,
					Name:        "ridz_",
					Subscriber:  true,
					Turbo:       false,
					ID:          "138035491",
					UserType:    "",
					VIP:         false,
				},
			},
		},
		{
			"@badge-info=;badges=;color=;display-name=Durrpadil;emotes=;first-msg=0;flags=;id=6b4dbb8a-b240-42d0-b890-d1f4be18cf10;mod=0;room-id=62463189;subscriber=0;tmi-sent-ts=1630887934441;turbo=0;user-id=39378800;user-type= :durrpadil!durrpadil@durrpadil.tmi.twitch.tv PRIVMSG #sophiabot :\u0001ACTION banned\u0001",
			PrivateMessage{
				Channel: "#sophiabot",
				IRCType: "PRIVMSG",
				Type:    PRIVMSG,
				Text:    "banned",
				Action:  true,
				Bits:    0,
				Emotes:  []Emote{},
				ID:      "6b4dbb8a-b240-42d0-b890-d1f4be18cf10",
				Reply:   false,
				User: &User{
					BadgeInfo:   "",
					Badges:      []Badge{},
					Broadcaster: false,
					Color:       "",
					DisplayName: "Durrpadil",
					Mod:         false,
					Name:        "durrpadil",
					Subscriber:  false,
					Turbo:       false,
					ID:          "39378800",
					UserType:    "",
					VIP:         false,
				},
			},
		},
	}

	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got := parsePrivateMessage(ircData)

		assertStringsEqual(t, "Channel", got.Channel, test.want.Channel)
		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		assertStringsEqual(t, "Text", got.Text, test.want.Text)
		assertBoolsEqual(t, "Action", got.Action, test.want.Action)
		assertIntsEqual(t, "Bits", got.Bits, test.want.Bits)
		assertEmoteSlicesEqual(t, got.Emotes, test.want.Emotes)
		assertStringsEqual(t, "ID", got.ID, test.want.ID)
		assertBoolsEqual(t, "Reply", got.Reply, test.want.Reply)
		assertUsersEqual(t, got.User, test.want.User)
	}
}

func TestParseWhisperMessage(t *testing.T) {
	tests := []struct {
		in   string
		want WhisperMessage
	}{
		{
			"@badges=;color=#FFFFFF;display-name=Bobby;emotes=25:37-41;message-id=1;thread-id=119302705_555556384;turbo=0;user-id=123456789;user-type= :bobby!bobby@bobby.tmi.twitch.tv WHISPER billy :hey look I'm a whisper with an emote Kappa",
			WhisperMessage{
				IRCType: "WHISPER",
				Type:    WHISPER,
				Text:    "hey look I'm a whisper with an emote Kappa",
				Emotes: []Emote{
					{"25", "Kappa", []EmotePosition{{37, 41}}},
				},
				ID:     "1",
				Target: "billy",
				User: &User{
					Badges:      []Badge{},
					Color:       "#FFFFFF",
					DisplayName: "Bobby",
					Name:        "bobby",
					Turbo:       false,
					ID:          "123456789",
					VIP:         false,
				},
			},
		},
		{
			"@badges=;color=#FFFF00;display-name=Boris;emotes=;message-id=2;thread-id=119302705_555556384;turbo=1;user-id=123456789;user-type= :boris!boris@boris.tmi.twitch.tv WHISPER bobby :a whisper without an emote",
			WhisperMessage{
				IRCType: "WHISPER",
				Type:    WHISPER,
				Text:    "a whisper without an emote",
				Emotes:  []Emote{},
				ID:      "2",
				Target:  "bobby",
				User: &User{
					Badges:      []Badge{},
					Color:       "#FFFF00",
					DisplayName: "Boris",
					Name:        "boris",
					Turbo:       true,
					ID:          "123456789",
					VIP:         false,
				},
			},
		},
	}

	for i := range tests {
		var test = tests[i]

		ircData, _ := parseIRCMessage(test.in)
		got := parseWhisperMessage(ircData)

		assertStringsEqual(t, "IRCType", got.IRCType, test.want.IRCType)
		assertMessageTypesEqual(t, got.Type, test.want.Type)
		assertStringsEqual(t, "Text", got.Text, test.want.Text)
		assertEmoteSlicesEqual(t, got.Emotes, test.want.Emotes)
		assertStringsEqual(t, "ID", got.ID, test.want.ID)
		assertStringsEqual(t, "Target", got.Target, test.want.Target)
		assertUsersEqual(t, got.User, test.want.User)
	}
}

func assertBoolsEqual(t *testing.T, name string, got, want bool) {
	if got != want {
		t.Errorf("%v: got %v, want %v", name, got, want)
	}
}

func assertDurationsEqual(t *testing.T, name string, got, want time.Duration) {
	if got != want {
		t.Errorf("%v: got %v, want %v", name, got, want)
	}
}

func assertIntsEqual(t *testing.T, name string, got, want int) {
	if got != want {
		t.Errorf("%v: got %v, want %v", name, got, want)
	}
}

func assertMessageTypesEqual(t *testing.T, got, want MessageType) {
	if got != want {
		t.Errorf("Type: got %v, want %v", got, want)
	}
}

func assertStringMapsEqual(t *testing.T, name string, got, want map[string]string) {
	if len(got) != len(want) {
		t.Errorf("%v: len(got) %v, len(want) %v", name, len(got), len(want))
	}
	for k, wv := range want {
		if gv, ok := got[k]; ok {
			if gv != wv {
				t.Errorf("%v[%v]: got %v, want %v", name, k, gv, wv)
			}
		} else {
			t.Errorf("%v[%v] not created in got, want %v, ", name, k, wv)
		}
	}
}

func assertStringsEqual(t *testing.T, name, got, want string) {
	if got != want {
		t.Errorf("%v: got %v, want %v", name, got, want)
	}
}

func assertStringSlicesEqual(t *testing.T, name string, got, want []string) {
	if len(got) != len(want) {
		t.Errorf("%v: len(got) %v, len(want) %v", name, len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%v[%v]: got %v, want %v", name, i, got[i], want[i])
		}
	}
}

func assertEmoteSlicesEqual(t *testing.T, got, want []Emote) {
	if len(got) != len(want) {
		t.Errorf("Emotes: len(got) %v, len(want) %v", len(got), len(want))
	}
	for i := range got {
		assertStringsEqual(t, "Emotes ID", got[i].ID, want[i].ID)
		assertStringsEqual(t, "Emotes Name", got[i].Name, want[i].Name)
		for j, p := range got[i].Positions {
			assertIntsEqual(t, fmt.Sprintf("Emotes Positions[%v].StartIdx", j), p.StartIdx, want[i].Positions[j].StartIdx)
			assertIntsEqual(t, fmt.Sprintf("Emotes Positions[%v].EndIdx", j), p.EndIdx, want[i].Positions[j].EndIdx)
		}
	}
}

func assertUsersEqual(t *testing.T, got, want *User) {
	assertStringsEqual(t, "BadgeInfo", got.BadgeInfo, want.BadgeInfo)
	if len(got.Badges) != len(want.Badges) {
		t.Errorf("Badges: len(got) %v, len(want) %v", len(got.Badges), len(want.Badges))
	}
	for i := range got.Badges {
		if got.Badges[i] != want.Badges[i] {
			t.Errorf("Badges[%v]: got %v, want %v", i, got.Badges[i], want.Badges[i])
		}
	}
	assertBoolsEqual(t, "Broadcaster", got.Broadcaster, want.Broadcaster)
	assertStringsEqual(t, "Color", got.Color, want.Color)
	assertStringsEqual(t, "DisplayName", got.DisplayName, want.DisplayName)
	assertBoolsEqual(t, "Mod", got.Mod, want.Mod)
	assertStringsEqual(t, "Name", got.Name, want.Name)
	assertBoolsEqual(t, "Subscriber", got.Subscriber, want.Subscriber)
	assertBoolsEqual(t, "Turbo", got.Turbo, want.Turbo)
	assertStringsEqual(t, "ID", got.ID, want.ID)
	assertStringsEqual(t, "UserType", got.UserType, want.UserType)
	assertBoolsEqual(t, "VIP", got.VIP, want.VIP)
}

func assertIRCDataEqual(t *testing.T, got, want *IRCData) {
	assertStringsEqual(t, "Raw", got.Raw, want.Raw)
	assertStringMapsEqual(t, "Tags", got.Tags, want.Tags)
	assertStringsEqual(t, "Prefix", got.Prefix, want.Prefix)
	assertStringsEqual(t, "Command", got.Command, want.Command)
	assertStringSlicesEqual(t, "Params", got.Params, want.Params)
}
