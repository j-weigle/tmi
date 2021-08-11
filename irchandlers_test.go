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
	client.OnDone(func(err error) {
		if err != nil {
			fmt.Println(err)
		} else {
			t.Errorf("client was supposed to error on login authentication")
		}
	})

	client.On(NOTICE, func(m Message) {
		var message, ok = m.(*NoticeMessage)
		if !ok {
			t.Errorf("Could not convert notice message to *NoticeMessage")
		}
		if message != nil {
			fmt.Println(message)
		} else {
			fmt.Println("message was nil")
		}
	})

	err := client.Connect()
	if err != ErrLoginFailure {
		t.Errorf("err type should have been fatal")
	}
}
