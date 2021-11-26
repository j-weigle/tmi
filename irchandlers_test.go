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
	config := NewClientConfig("", "")
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

	config := NewClientConfig("", "")
	config.Identity.Anonymous()
	config.Pinger.Enabled = false
	c := NewClient(config)

	rcvdPong := make(chan struct{})
	c.OnPongMessage(func(m PongMessage) {
		rcvdPong <- struct{}{}
	})

	go func() {
		err := c.Connect()
		if err == nil {
			t.Errorf("Connect should always return an error when done")
		}
	}()
	time.Sleep(time.Second)
	c.send("PING :" + PingSignature)
	select {
	case <-rcvdPong:
	case <-time.After(time.Second * 5):
		t.Errorf("expected pong after 5 seconds")
	}
	c.send("PING")
	select {
	case <-rcvdPong:
	case <-time.After(time.Second * 5):
		t.Errorf("expected pong after 5 seconds")
	}

	close(rcvdPong)
	c.Disconnect()
}

func TestHandleIRCMessage(t *testing.T) {
	tests := []struct {
		in   string
		want error
	}{
		{"001", nil},
		{"CLEARCHAT", nil},
		{"CLEARMSG", nil},
		{"GLOBALUSERSTATE", nil},
		{"HOSTTARGET", nil},
		{"NOTICE", nil},
		{"NOTICE * :Login authentication failed", ErrLoginFailure},
		{"RECONNECT", errReconnect},
		{"ROOMSTATE", nil},
		{"USERNOTICE", nil},
		{"USERSTATE", nil},
		{"353", nil},
		{"JOIN", nil},
		{"PART", nil},
		{"PING", nil},
		{"PONG", nil},
		{"PRIVMSG", nil},
		{"WHISPER", nil},
		{"002", nil},
		{"003", nil},
		{"004", nil},
		{"375", nil},
		{"372", nil},
		{"376", nil},
		{"CAP", nil},
		{"SERVERCHANGE", nil},
		{"421", nil},
		{"MODE", nil},
		{"366", nil},
	}

	c := NewClient(NewClientConfig("", ""))
	for _, test := range tests {
		got := c.handleIRCMessage(test.in)
		if got != test.want {
			t.Errorf("got error: %v, want error: %v", got, test.want)
		}
	}
}

func TestTmiHandlers(t *testing.T) {
	tests := []struct {
		inRaw string
		want  error
	}{
		{"001", nil},
		{"CLEARCHAT", nil},
		{"CLEARMSG", nil},
		{"GLOBALUSERSTATE", nil},
		{"HOSTTARGET", nil},
		{"NOTICE", nil},
		{"NOTICE * :Login authentication failed", ErrLoginFailure},
		{"RECONNECT", errReconnect},
		{"ROOMSTATE", nil},
		{"USERNOTICE", nil},
		{"USERSTATE", nil},
		{"353", nil},
		{"JOIN", nil},
		{"PART", nil},
		{"PING", nil},
		{"PONG", nil},
		{"PRIVMSG", nil},
		{"WHISPER", nil},
		{"002", errUnsetIRCCommand},
		{"003", errUnsetIRCCommand},
		{"004", errUnsetIRCCommand},
		{"375", errUnsetIRCCommand},
		{"372", errUnsetIRCCommand},
		{"376", errUnsetIRCCommand},
		{"CAP", errUnsetIRCCommand},
		{"SERVERCHANGE", errUnsetIRCCommand},
		{"421", errUnsetIRCCommand},
		{"MODE", errUnsetIRCCommand},
		{"366", errUnsetIRCCommand},
		{"RANDOMCOMMAND", errUnrecognizedIRCCommand},
	}

	c := NewClient(NewClientConfig("", ""))
	for _, test := range tests {
		in, err := parseIRCMessage(test.inRaw)
		if err != nil {
			t.Error(err)
		}
		got := c.handleIRCData(in)
		if got != test.want {
			t.Errorf("got error: %v, want error: %v", got, test.want)
		}
	}
}

func TestAllHandlersCallOnMessageWhenSet(t *testing.T) {
	results := make(map[MessageType]bool)
	types := []MessageType{UNSET, CLEARCHAT, CLEARMSG, GLOBALUSERSTATE, HOSTTARGET, NOTICE, RECONNECT, ROOMSTATE, USERNOTICE, USERSTATE, NAMES, JOIN, PART, PING, PONG, PRIVMSG, WHISPER}

	var wantUnsetCounter = 12
	var unsetCounter int
	var onConnectedCalled bool

	tests := []struct {
		in   string
		want error
	}{
		{"001", nil},
		{"CLEARCHAT", nil},
		{"CLEARMSG", nil},
		{"GLOBALUSERSTATE", nil},
		{"HOSTTARGET", nil},
		{"NOTICE", nil},
		{"NOTICE * :Login authentication failed", ErrLoginFailure},
		{"RECONNECT", errReconnect},
		{"ROOMSTATE", nil},
		{"USERNOTICE", nil},
		{"USERSTATE", nil},
		{"353", nil},
		{"JOIN", nil},
		{"PART", nil},
		{"PING", nil},
		{"PONG", nil},
		{"PRIVMSG", nil},
		{"WHISPER", nil},
		{"002", nil},
		{"003", nil},
		{"004", nil},
		{"375", nil},
		{"372", nil},
		{"376", nil},
		{"CAP", nil},
		{"SERVERCHANGE", nil},
		{"421", nil},
		{"MODE", nil},
		{"366", nil},
		{"RANDOMCOMMAND", nil},
	}

	c := NewClient(NewClientConfig("", ""))

	c.OnUnsetMessage(func(m UnsetMessage) { results[m.Type] = true; unsetCounter++ })
	c.OnConnected(func() { onConnectedCalled = true })
	c.OnClearChatMessage(func(m ClearChatMessage) { results[m.Type] = true })
	c.OnClearMsgMessage(func(m ClearMsgMessage) { results[m.Type] = true })
	c.OnGlobalUserstateMessage(func(m GlobalUserstateMessage) { results[m.Type] = true })
	c.OnHostTargetMessage(func(m HostTargetMessage) { results[m.Type] = true })
	c.OnNoticeMessage(func(m NoticeMessage) { results[m.Type] = true })
	c.OnReconnectMessage(func(m ReconnectMessage) { results[m.Type] = true })
	c.OnRoomstateMessage(func(m RoomstateMessage) { results[m.Type] = true })
	c.OnUserNoticeMessage(func(m UsernoticeMessage) { results[m.Type] = true })
	c.OnUserstateMessage(func(m UserstateMessage) { results[m.Type] = true })
	c.OnNamesMessage(func(m NamesMessage) { results[m.Type] = true })
	c.OnJoinMessage(func(m JoinMessage) { results[m.Type] = true })
	c.OnPartMessage(func(m PartMessage) { results[m.Type] = true })
	c.OnPingMessage(func(m PingMessage) { results[m.Type] = true })
	c.OnPongMessage(func(m PongMessage) { results[m.Type] = true })
	c.OnPrivateMessage(func(m PrivateMessage) { results[m.Type] = true })
	c.OnWhisperMessage(func(m WhisperMessage) { results[m.Type] = true })

	for _, test := range tests {
		err := c.handleIRCMessage(test.in)
		if err != test.want {
			t.Error(err)
		}
	}

	for _, mt := range types {
		if r, ok := results[mt]; ok {
			if !r {
				t.Errorf("%v MessageType handler not called", mt)
			}
		}
	}
	if unsetCounter != wantUnsetCounter {
		t.Errorf("number of Unset MessageType handler calls %v, want %v", unsetCounter, wantUnsetCounter)
	}
	if !onConnectedCalled {
		t.Errorf("OnConnect handler never called")
	}
}
