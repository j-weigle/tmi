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

// 421(INVALIDIRC) message handler
func TestTmiTwitchTvCommand421(t *testing.T) {
	var ircType = "421"
	var username = "ausername"
	var unknown = "WHO"
	var text = "Unknown command"

	var testMessage = fmt.Sprintf(":tmi.twitch.tv %v %v %v :%v", ircType, username, unknown, text)

	var want = InvalidIRCMessage{
		IRCType: ircType,
		Text:    text,
		Type:    INVALIDIRC,
		Unknown: unknown,
		User:    username,
	}

	config := NewClientConfig()
	c := NewClient(config)
	c.OnInvalidIRCMessage(func(message InvalidIRCMessage) {
		message.Data = IRCData{}
		if !want.Data.equals(&message.Data) {
			t.Errorf("want: %+v, got : %+v", want, message)
		}
		if want.IRCType != message.IRCType {
			t.Errorf("want: %+v, got : %+v", want, message)
		}
		if want.Text != message.Text {
			t.Errorf("want: %+v, got : %+v", want, message)
		}
		if want.Type != message.Type {
			t.Errorf("want: %+v, got : %+v", want, message)
		}
		if want.Unknown != message.Unknown {
			t.Errorf("want: %+v, got : %+v", want, message)
		}
		if want.User != message.User {
			t.Errorf("want: %+v, got : %+v", want, message)
		}
	})

	var ircData, _ = parseIRCMessage(testMessage)
	var err = c.tmiTwitchTvCommand421(ircData)
	if err != nil {
		t.Error(err)
	}
}
