package tmi

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestAnonymousConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skip:TestAnonymousConnection, reason:short mode")
	}

	config := NewClientConfig("", "")
	config.Identity.Anonymous()
	client := NewClient(config)

	var errConnecting error = fmt.Errorf("anonymous client connection failed")
	client.OnConnected(func() {
		errConnecting = nil
	})

	go func() {
		time.Sleep(time.Second)
		client.Disconnect()
	}()
	err := client.Connect()
	if err != ErrDisconnectCalled {
		t.Error(err)
	}
	if errConnecting != nil {
		t.Error(errConnecting)
	}
}

func TestJoinThreeChannels(t *testing.T) {
	if testing.Short() {
		t.Skip("skip:TestJoinThreeChannels, reason:short mode")
	}
	tests := map[string]bool{
		"#twitch":      false,
		"#testchannel": false,
		"#twitchmedia": false,
	}
	channels := make([]string, 0, len(tests))
	for ch := range tests {
		channels = append(channels, ch)
	}

	config := NewClientConfig("", "")
	config.Identity.Anonymous()
	c := NewClient(config)

	c.OnJoinMessage(func(m JoinMessage) {
		tests[m.Channel] = true
	})

	err := c.Join(channels...)
	if err != nil {
		t.Error(err)
	}

	go func() {
		// give it a second to join the channels and receive the join messages
		time.Sleep(time.Second)
		for ch := range tests {
			if joined := tests[ch]; !joined {
				t.Errorf("did not join %v", ch)
			}
			err := c.Part(ch)
			if err != nil {
				t.Error(err)
			}
		}
		c.Disconnect()
	}()
	err = c.Connect()
	if err == nil {
		t.Errorf("Connect should always return an error when done")
	}
}

func TestPartChannels(t *testing.T) {
	if testing.Short() {
		t.Skip("skip:TestPartChannels, reason:short mode")
	}
	tests := map[string]bool{
		"#twitch":      false,
		"#testchannel": false,
	}
	channels := make([]string, 0, len(tests))
	for ch := range tests {
		channels = append(channels, ch)
	}

	config := NewClientConfig("", "")
	config.Identity.Anonymous()
	c := NewClient(config)

	c.OnJoinMessage(func(m JoinMessage) {
		tests[m.Channel] = true
	})

	c.OnPartMessage(func(m PartMessage) {
		tests[m.Channel] = false
	})

	err := c.Join(channels...)
	if err != nil {
		t.Error(err)
	}

	go func() {
		// give it a second to join the channels and receive the join messages
		time.Sleep(time.Second)
		for ch := range tests {
			if joined := tests[ch]; !joined {
				t.Errorf("did not join %v", ch)
			}
			err := c.Part(ch)
			if err != nil {
				t.Error(err)
			}
			// give it a second to part the channel and receive the part message
			time.Sleep(time.Second)
			if joined := tests[ch]; joined {
				t.Errorf("did not part %v", ch)
			}
		}
		c.Disconnect()
	}()
	err = c.Connect()
	if err == nil {
		t.Errorf("Connect should always return an error when done")
	}
}

func TestFormatChannel(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"test", "#test"},
		{"chEcking", "#checking"},
		{"#billY", "#billy"},
		{"#bobby", "#bobby"},
		{" oops", "#oops"},
	}

	for _, test := range tests {
		got := formatChannel(test.in)
		if got != test.want {
			t.Errorf("got %v, want %v", got, test.want)
		}
	}
}

func TestSay(t *testing.T) {
	type testInput struct {
		channel string
		message string
	}
	tests := []struct {
		in   testInput
		want string
	}{
		{testInput{"test", "hello test"}, "PRIVMSG #test :hello test"},
		{testInput{"#Checking", "I am checking"}, "PRIVMSG #checking :I am checking"},
		{testInput{"#beep", "boop"}, "PRIVMSG #beep :boop"},
		{testInput{"CHANNEL", "yeah okay"}, "PRIVMSG #channel :yeah okay"},
	}

	c := NewClient(NewClientConfig("", ""))
	for _, test := range tests {
		c.Say(test.in.channel, test.in.message)
		got := <-c.outbound
		if got != test.want {
			t.Errorf("got %v, want %v", got, test.want)
		}
	}
}

func TestSayLong(t *testing.T) {
	test := strings.Repeat("x ", 510)
	wants := []string{"PRIVMSG #long :" + strings.TrimSpace(test[:500]),
		"PRIVMSG #long :" + strings.TrimSpace(test[500:1000]),
		"PRIVMSG #long :" + strings.TrimSpace(test[1000:])}

	c := NewClient(NewClientConfig("", ""))
	c.Say("#long", test)

	for _, want := range wants {
		got := <-c.outbound
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func TestUpdatePassword(t *testing.T) {
	want := "oauth:newpassword"
	pw := "newpassword"
	c := NewClient(NewClientConfig("", ""))
	c.UpdatePassword(pw)
	if c.config.Identity.Password != want {
		t.Errorf("got %v, want %v", c.config.Identity.Password, want)
	}
}

func TestAction(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{
			"",
			"PRIVMSG #channel :\u0001ACTION \u0001",
		},
		{
			"jumps for joy",
			"PRIVMSG #channel :\u0001ACTION jumps for joy\u0001",
		},
	}

	c := NewClient(NewClientConfig("", ""))
	for i := range tests {
		var test = tests[i]
		err := c.Action("#channel", test.in)
		if err != nil {
			t.Error(err)
		}
		got := <-c.outbound
		if got != test.want {
			t.Errorf("got %v, want %v", got, test.want)
		}
	}

	err := c.Action("#channel", strings.Repeat("x", 491))
	if err == nil {
		t.Errorf("expected message too long error")
	}
}

func TestBan(t *testing.T) {
	type UserReason struct {
		User   string
		Reason string
	}
	tests := []struct {
		in   UserReason
		want string
	}{
		{
			UserReason{"bobby", "being rude"},
			"PRIVMSG #channel :/ban bobby being rude",
		},
		{
			UserReason{"billy", ""},
			"PRIVMSG #channel :/ban billy",
		},
	}

	c := NewClient(NewClientConfig("", ""))
	for i := range tests {
		var test = tests[i]
		err := c.Ban("#channel", test.in.User, test.in.Reason)
		if err != nil {
			t.Error(err)
		}
		got := <-c.outbound
		if got != test.want {
			t.Errorf("got %v, want %v", got, test.want)
		}
	}
	err := c.Ban("#channel", "anyone", strings.Repeat("x", 491))
	if err == nil {
		t.Errorf("expected message too long error")
	}
}

func TestUnban(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Unban("#channel", "user")
	want := "PRIVMSG #channel :/unban user"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestClear(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Clear("#channel")
	want := "PRIVMSG #channel :/clear"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestColor(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Color("#AABBCC")
	want := "PRIVMSG # :/color #AABBCC"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	c.config.Identity.Username = "name"
	c.Color("#AABBCC")
	want = "PRIVMSG #name :/color #AABBCC"
	got = <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestCommercial(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Commercial("#channel", "30")
	want := "PRIVMSG #channel :/commercial 30"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDelete(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Delete("#channel", "1234-5678")
	want := "PRIVMSG #channel :/delete 1234-5678"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestEmoteOnly(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.EmoteOnly("#channel")
	want := "PRIVMSG #channel :/emoteonly"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestEmoteOnlyOff(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.EmoteOnlyOff("#channel")
	want := "PRIVMSG #channel :/emoteonlyoff"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFollowers(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Followers("#channel", "10m")
	want := "PRIVMSG #channel :/followers 10m"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFollowersOff(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.FollowersOff("#channel")
	want := "PRIVMSG #channel :/followersoff"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestHost(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Host("#channel", "#target")
	want := "PRIVMSG #channel :/host target"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestUnhost(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Unhost("#channel")
	want := "PRIVMSG #channel :/unhost"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarker(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	err := c.Marker("#channel", "")
	if err != nil {
		t.Error(err)
	}
	want := "PRIVMSG #channel :/marker"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	err = c.Marker("#channel", "test")
	if err != nil {
		t.Error(err)
	}
	want = "PRIVMSG #channel :/marker test"
	got = <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMod(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Mod("#channel", "user")
	want := "PRIVMSG #channel :/mod user"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestUnmod(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Unmod("#channel", "user")
	want := "PRIVMSG #channel :/unmod user"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMods(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Mods("#channel")
	want := "PRIVMSG #channel :/mods"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestR9kBetaAndAliases(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.R9kBeta("#channel")
	c.R9kMode("#channel")
	c.Uniquechat("#channel")
	want := "PRIVMSG #channel :/r9kbeta"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	got = <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	got = <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestR9kBetaOffAndAliases(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.R9kBetaOff("#channel")
	c.R9kModeOff("#channel")
	c.UniquechatOff("#channel")
	want := "PRIVMSG #channel :/r9kbetaoff"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	got = <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	got = <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRaid(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Raid("#channel", "target")
	want := "PRIVMSG #channel :/raid target"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestUnraid(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Unraid("#channel")
	want := "PRIVMSG #channel :/unraid"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSlow(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Slow("#channel", "3")
	want := "PRIVMSG #channel :/slow 3"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSlowOff(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.SlowOff("#channel")
	want := "PRIVMSG #channel :/slowoff"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSubscribers(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Subscribers("#channel")
	want := "PRIVMSG #channel :/subscribers"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSubscribersOff(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.SubscribersOff("#channel")
	want := "PRIVMSG #channel :/subscribersoff"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestTimeout(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Timeout("#channel", "user", "")
	want := "PRIVMSG #channel :/timeout user"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	c.Timeout("#channel", "user", "30")
	want = "PRIVMSG #channel :/timeout user 30"
	got = <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestUntimeout(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.Untimeout("#channel", "user")
	want := "PRIVMSG #channel :/untimeout user"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestVIP(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.VIP("#channel", "user")
	want := "PRIVMSG #channel :/vip user"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestUnVIP(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.UnVIP("#channel", "user")
	want := "PRIVMSG #channel :/unvip user"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestVIPs(t *testing.T) {
	c := NewClient(NewClientConfig("", ""))
	c.VIPs("#channel")
	want := "PRIVMSG #channel :/vips"
	got := <-c.outbound
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
