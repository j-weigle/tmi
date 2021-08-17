package tmi

import (
	"fmt"
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
	client.OnDone(func(err error) {
		if err != ErrLoginFailure {
			t.Errorf("client was supposed to error on login authentication")
		} else {
			fmt.Println(err)
		}
	})

	client.OnNoticeMessage(func(message NoticeMessage) {
		fmt.Println(message.Text)
	})

	err := client.Connect()
	if err != ErrLoginFailure {
		t.Errorf("client was supposed to error on login authentication")
	}
}
