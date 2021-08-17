package tmi

import (
	"testing"
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
