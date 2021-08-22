package tmi

import (
	"context"
	"testing"
)

func TestListenAndParse(t *testing.T) {
	var c = NewClient(NewClientConfig())

	// Disconnect called
	var ctx, cancelFunc = context.WithCancel(context.Background())
	c.notifDisconnect.reset()
	var closeErr = connCloseErr{}
	var closeErrCb = func(errReason error) {
		closeErr.update(errReason)
	}

	c.notifDisconnect.notify()

	c.listenAndParse(ctx, closeErrCb)

	if closeErr.err != ErrDisconnectCalled {
		t.Errorf("expected error: %v, got error: %v", ErrDisconnectCalled, closeErr.err)
	}

	// RECONNECT message received
	c.notifDisconnect.reset()
	closeErr = connCloseErr{}
	closeErrCb = func(errReason error) {
		closeErr.update(errReason)
	}

	c.inbound <- ":tmi.twitch.tv RECONNECT"

	c.listenAndParse(ctx, closeErrCb)

	if closeErr.err != errReconnect {
		t.Errorf("expected error: %v, got error: %v", errReconnect, closeErr.err)
	}

	// Context closed
	closeErr = connCloseErr{}
	closeErrCb = func(errReason error) {
		closeErr.update(errReason)
	}

	cancelFunc()

	c.listenAndParse(ctx, closeErrCb)

	if closeErr.err != nil {
		t.Errorf("expected error: nil, got error: %v", closeErr.err)
	}
}
