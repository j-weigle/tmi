package tmi

import (
	"testing"
	"time"
)

// NOTICE message handler
func TestLoginFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skip:TestLoginFailure, reason:short mode")
	}
	config := NewClientConfig()
	config.Identity.Set("a", "blah")

	client := NewClient(config)

	err := client.Connect()
	if err != ErrLoginFailure {
		t.Errorf("client was supposed to error on login authentication")
	}
}

func TestPingPong(t *testing.T) {
	if testing.Short() {
		t.Skip("skip:TestPingPong, reason:short mode")
	}

	config := NewClientConfig()
	config.Identity.Anonymous()
	config.Pinger.Disable()
	c := NewClient(config)

	rcvdPong := make(chan struct{})
	c.OnPongMessage(func(m PongMessage) {
		rcvdPong <- struct{}{}
	})

	go c.Connect()
	time.Sleep(time.Second)
	err := c.send("PING :" + pingSignature)
	if err != nil {
		t.Error(err)
	}
	select {
	case <-rcvdPong:
	case <-time.After(time.Second * 5):
		t.Errorf("expected pong after 5 seconds")
	}
	err = c.send("PING")
	if err != nil {
		t.Error(err)
	}
	select {
	case <-rcvdPong:
	case <-time.After(time.Second * 5):
		t.Errorf("expected pong after 5 seconds")
	}

	close(rcvdPong)
	c.Disconnect()
}
