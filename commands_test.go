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

	config := NewClientConfig()
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

	config := NewClientConfig()
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
			c.Part(ch)
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

	config := NewClientConfig()
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

	c := NewClient(NewClientConfig())
	for _, test := range tests {
		err := c.Say(test.in.channel, test.in.message)
		if err != nil {
			t.Error(err)
		}
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

	c := NewClient(NewClientConfig())
	err := c.Say("#long", test)
	if err != nil {
		t.Error(err)
	}

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
	c := NewClient(NewClientConfig())
	c.UpdatePassword(pw)
	if c.config.Identity.password != want {
		t.Errorf("got %v, want %v", c.config.Identity.password, want)
	}
}
