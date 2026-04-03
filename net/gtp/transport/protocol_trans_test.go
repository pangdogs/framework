package transport

import (
	"testing"

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/net/gtp"
)

func TestTransProtocolSendDataAndHandleEvent(t *testing.T) {
	client, server := newUnsequencedPipeTransceivers(t)
	trans := &TransProtocol{Transceiver: client}

	done := make(chan error, 1)
	go func() { done <- trans.SendData([]byte("hello")) }()
	e := recvWithTimeout(t, server)
	if err := <-done; err != nil {
		t.Fatalf("SendData failed: %v", err)
	}
	if string(AssertEvent[*gtp.MsgPayload](e).Msg.Data) != "hello" {
		t.Fatalf("unexpected payload event: %+v", e)
	}

	called := false
	trans.PayloadHandler = generic.CastDelegateVoid1(func(e Event[*gtp.MsgPayload]) {
		called = string(e.Msg.Data) == "world"
	})
	trans.HandleEvent(Event[*gtp.MsgPayload]{Msg: &gtp.MsgPayload{Data: []byte("world")}}.Interface())
	if !called {
		t.Fatal("expected payload handler")
	}
}

func TestTransProtocolNilTransceiver(t *testing.T) {
	if err := (&TransProtocol{}).SendData([]byte("x")); err == nil {
		t.Fatal("expected SendData error")
	}
}
