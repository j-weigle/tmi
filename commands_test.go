package tmi

import (
	"fmt"
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
