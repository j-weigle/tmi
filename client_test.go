package tmi

import (
	"context"
	"testing"
)

func TestReadInbound(t *testing.T) {
	var client = NewClient(NewClientConfig())

	// Disconnect called
	var ctx, cancelFunc = context.WithCancel(context.Background())
	client.notifDisconnect.reset()
	var closeErr = connCloseErr{}
	var closeErrCb = func(errReason error) {
		closeErr.update(errReason)
	}

	client.notifDisconnect.notify()

	client.readInbound(ctx, closeErrCb)

	if closeErr.err != ErrDisconnectCalled {
		t.Errorf("expected error: %v, got error: %v", ErrDisconnectCalled, closeErr.err)
	}

	// RECONNECT message received
	client.notifDisconnect.reset()
	closeErr = connCloseErr{}
	closeErrCb = func(errReason error) {
		closeErr.update(errReason)
	}

	client.inbound <- ":tmi.twitch.tv RECONNECT"

	client.readInbound(ctx, closeErrCb)

	if closeErr.err != errReconnect {
		t.Errorf("expected error: %v, got error: %v", errReconnect, closeErr.err)
	}

	closeErr = connCloseErr{}
	closeErrCb = func(errReason error) {
		closeErr.update(errReason)
	}

	cancelFunc()

	client.readInbound(ctx, closeErrCb)

	if closeErr.err != nil {
		t.Errorf("expected error: nil, got error: %v", closeErr.err)
	}
}
