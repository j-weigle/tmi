package tmi

import (
	"fmt"
	"testing"
	"time"
)

func TestLoginFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skip:TestLoginFailure, reason:short mode")
	}
	config := NewClientConfig()
	config.Identity.Set("a", "blah")
	config.Connection.SetSync(true)

	client := NewClient(config)
	client.Done(func() {
		var err = client.Failure()
		if err != nil {
			fmt.Println(err)
		} else {
			t.Errorf("client was supposed to error on login authentication")
		}
	})

	err := client.Connect()
	if err != nil {
		t.Errorf("connection error occurred instead of login error")
	}
}

func TestDoubleConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skip:TestDoubleConnect, reason:short mode")
	}
	config := NewClientConfig()
	config.Identity.Anonymous()
	config.Channels = []string{"twitch"}

	client := NewClient(config)

	err := client.Connect()
	if err != nil {
		t.Errorf("failed on initial connection:")
		t.Error(err)
	}
	// attempt to reconnect after 5 seconds, then disconnect
	time.Sleep(time.Second * 5)
	err = client.Connect()
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 5)
	err = client.Disconnect()
	if err != nil {
		t.Error(err)
	}
}
