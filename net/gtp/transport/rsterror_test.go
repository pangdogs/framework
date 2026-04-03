package transport

import (
	"testing"

	"git.golaxy.org/framework/net/gtp"
)

func TestRstErrorConversions(t *testing.T) {
	err := RstError{Code: gtp.Code_Reject, Message: "bad request"}
	if got := err.Error(); got == "" {
		t.Fatal("expected error string")
	}

	event := err.ToEvent()
	if event.Msg.Code != err.Code || event.Msg.Message != err.Message {
		t.Fatalf("unexpected rst event: %+v", event)
	}

	cloned := CastRstErr(event)
	if cloned.Code != err.Code || cloned.Message != err.Message {
		t.Fatalf("unexpected cloned rst error: %+v", cloned)
	}
}
