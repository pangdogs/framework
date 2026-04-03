package transport

import (
	"context"
	"testing"

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/net/gtp"
)

func TestEventDispatcherDispatch(t *testing.T) {
	client, server := newUnsequencedPipeTransceivers(t)
	done := make(chan error, 1)
	go func() { done <- client.Send(newPayloadEvent("dispatch")) }()

	called := false
	dispatcher := EventDispatcher{
		Transceiver: server,
		EventHandler: generic.CastDelegateVoid1(func(e IEvent) {
			called = string(AssertEvent[*gtp.MsgPayload](e).Msg.Data) == "dispatch"
		}),
	}

	if err := dispatcher.Dispatch(context.Background()); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if err := <-done; err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
}

func TestEventDispatcherNilTransceiver(t *testing.T) {
	if err := (&EventDispatcher{}).Dispatch(context.Background()); err == nil {
		t.Fatal("expected Dispatch error")
	}
}
