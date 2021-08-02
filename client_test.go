package tmi

import (
	"strconv"
	"testing"
)

func TestNewClientConfig(t *testing.T) {
	connection := &connectionConfig{true, true, 1000, -1, 30000}
	id := &identityConfig{}
	channels := []string{}

	want := &clientConfig{connection, id, channels}
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
	config.Identity.SetToAnonymous()

	jf := config.Identity.username[:9]
	if jf != "justinfan" {
		t.Errorf("SetToAnonymous should make username start with justinfan")
	}
	_, err := strconv.Atoi(config.Identity.username[10:])
	if err != nil {
		t.Errorf("SetToAnonymous should generate a random integer to end justinfan username")
	}
}
