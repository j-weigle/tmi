package tmi

import (
	"strconv"
	"testing"
	"time"
)

func TestNewClientConfig(t *testing.T) {
	connection := &connectionConfig{true, true, -1, 30000}
	id := &identityConfig{}
	pinger := &pingConfig{
		wait:    time.Second * 60,
		timeout: time.Second * 5,
	}
	channels := []string{}

	want := &clientConfig{channels, connection, id, pinger}
	got := NewClientConfig()

	if *want.Connection != *got.Connection {
		t.Errorf("NewClientConfig().Connection == %v, want %v", got.Connection, want.Connection)
	}
	if *want.Identity != *got.Identity {
		t.Errorf("NewClientConfig().Identity == %v, want %v", got.Identity, want.Identity)
	}
	if len(want.Channels) == len(got.Channels) {
		for i, w := range want.Channels {
			if w != got.Channels[i] {
				t.Errorf("NewClientConfig().Channels == %v, want %v", got.Channels, want.Channels)
			}
		}
	} else {
		t.Errorf("NewClientConfig().Channels == %v, want %v", got.Channels, want.Channels)
	}
}

func TestSetToAnonymous(t *testing.T) {
	config := NewClientConfig()
	config.Identity.Anonymous()

	jf := config.Identity.username[:9]
	if jf != "justinfan" {
		t.Errorf("SetToAnonymous should make username start with justinfan")
	}
	_, err := strconv.Atoi(config.Identity.username[10:])
	if err != nil {
		t.Errorf("SetToAnonymous should generate a random integer to end justinfan username")
	}
}
