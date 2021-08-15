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

	client.On(NOTICE, func(m Message) {
		var message, ok = m.(*NoticeMessage)
		if !ok {
			t.Errorf("Could not convert notice message to *NoticeMessage")
		}
		if message != nil {
			fmt.Println(message.Text)
		} else {
			t.Errorf("notice message was nil")
		}
	})

	err := client.Connect()
	if err != ErrLoginFailure {
		t.Errorf("client was supposed to error on login authentication")
	}
}

// 421(INVALIDIRC) message handler
func TestTmiTwitchTvCommand421(t *testing.T) {
	var ircType = "421"
	var username = "ausername"
	var unknown = "WHO"
	var text = "Unknown command"

	var testMessage = fmt.Sprintf(":tmi.twitch.tv %v %v %v :%v", ircType, username, unknown, text)

	var want = &InvalidIRCMessage{
		IRCType: ircType,
		Text:    text,
		Type:    INVALIDIRC,
		Unknown: unknown,
		User:    username,
	}

	config := NewClientConfig()
	c := NewClient(config)
	c.On(INVALIDIRC, func(m Message) {
		var message, ok = m.(*InvalidIRCMessage)
		if !ok {
			t.Errorf("Could not convert notice message to *InvalidIRCMessage")
		}
		if message != nil {
			message.Data = nil
			if *want != *message {
				t.Errorf("want: %+v, got : %+v", want, message)
			}
		} else {
			t.Errorf("421 message was nil")
		}
	})

	var client = c.(*client) // convert Client interface to *client struct

	var ircData, _ = parseIRCMessage(testMessage)
	var err = client.tmiTwitchTvCommand421(ircData)
	if err != nil {
		t.Error(err)
	}
}
