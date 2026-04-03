package transport

import (
	"testing"

	"git.golaxy.org/framework/net/gtp"
)

func TestEventInterfaceAndAssertEvent(t *testing.T) {
	e := Event[*gtp.MsgPayload]{
		Flags: gtp.Flags_None().Setd(gtp.Flag_Compressed, true),
		Seq:   1,
		Ack:   2,
		Msg:   &gtp.MsgPayload{Data: []byte("hello")},
	}

	i := e.Interface()
	got := AssertEvent[*gtp.MsgPayload](i)
	if got.Seq != e.Seq || got.Ack != e.Ack || string(got.Msg.Data) != "hello" {
		t.Fatalf("unexpected asserted event: %+v", got)
	}
}

func TestAssertEventPanicsOnIncorrectType(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	_ = AssertEvent[*gtp.MsgHeartbeat](newPayloadEvent("x"))
}
