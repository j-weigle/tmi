package tmi

import (
	"fmt"
	"testing"
)

func TestLoginFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skip:TestLoginFailure, reason:short mode")
	}
	config := NewClientConfig()
	config.Identity.Set("a", "blah")

	client := NewClient(config)
	client.Done(func(err error) {
		if err != nil {
			fmt.Println(err)
		} else {
			t.Errorf("client was supposed to error on login authentication")
		}
	})

	err := client.Connect()
	if err != errFatalNotification {
		t.Errorf("err type should have been fatal")
	}
}
