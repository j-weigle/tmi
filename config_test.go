package tmi

import (
	"strconv"
	"testing"
	"time"
)

func TestNewClientConfig(t *testing.T) {
	connection := ConnectionConfig{true, true, -1, time.Second * 30}
	id := IdentityConfig{}
	pinger := PingConfig{true, time.Minute, time.Second * 5}

	want := &ClientConfig{connection, id, pinger}
	got := NewClientConfig("", "")

	if want.Connection != got.Connection {
		t.Errorf("Connection: got %v, want %v", got.Connection, want.Connection)
	}
	if want.Identity != got.Identity {
		t.Errorf("Identity: got %v, want %v", got.Identity, want.Identity)
	}
	if want.Pinger != got.Pinger {
		t.Errorf("Pinger: got %v, want %v", got.Pinger, want.Pinger)
	}
}

func TestAnonymous(t *testing.T) {
	config := NewClientConfig("", "")
	config.Identity.Anonymous()

	jf := config.Identity.Username[:9]
	if jf != "justinfan" {
		t.Errorf("SetToAnonymous should make username start with justinfan")
	}
	_, err := strconv.Atoi(config.Identity.Username[10:])
	if err != nil {
		t.Errorf("SetToAnonymous should generate a random integer to end justinfan username")
	}
}

func TestSetReconnectSettings(t *testing.T) {
	config := NewClientConfig("", "")
	config.Connection.SetReconnectSettings(20, time.Second*6)

	if config.Connection.MaxReconnectAttempts != 20 {
		t.Errorf("maxReconnectAttempts: got %v, want %v", config.Connection.MaxReconnectAttempts, 20)
	}
	if config.Connection.MaxReconnectInterval != time.Second*6 {
		t.Errorf("maxReconnectInterval: got %v, want %v", config.Connection.MaxReconnectInterval, time.Second*6)
	}

	config.Connection.SetReconnectSettings(20, time.Second*4)
	if config.Connection.MaxReconnectInterval != time.Second*5 {
		t.Errorf("maxReconnectInterval: got %v, want %v", config.Connection.MaxReconnectInterval, time.Second*5)
	}
}

func TestSetPassword(t *testing.T) {
	config := NewClientConfig("", "")
	config.Identity.SetPassword("p")
	var want = "oauth:p"

	if config.Identity.Password != want {
		t.Errorf("SetPassword: got %v, want %v", config.Identity.Password, want)
	}
}
