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
