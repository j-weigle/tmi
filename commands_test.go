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
	tests := []string{"#twitch", "#testchannel", "#twitchmedia"}
	results := make(map[string]bool)
	for _, test := range tests {
		results[test] = false
	}

	config := NewClientConfig()
	config.Identity.Anonymous()
	c := NewClient(config)

	c.OnJoinMessage(func(m JoinMessage) {
		results[m.Channel] = true
	})

	c.Join(tests...)

	go func() {
		// join waits 600 milliseconds between joins, give it extra time
		time.Sleep(time.Millisecond * 650 * 3)
		c.Disconnect()
	}()
	c.Connect()

	for _, test := range tests {
		if joined, ok := results[test]; !(ok && joined) {
			t.Errorf("did not join %v", test)
		}
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

	c := NewClient(NewClientConfig())
	c.Say("#long", test)

	for _, want := range wants {
		got := <-c.outbound
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}
