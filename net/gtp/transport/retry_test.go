package transport

import (
	"context"
	"errors"
	"io"
	"testing"

	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/codec"
)

var retryErrTest = errors.New("retry test")

func TestRetrySend(t *testing.T) {
	if err := (Retry{}).Send(nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := (Retry{}).Send(retryErrTest); !errors.Is(err, retryErrTest) {
		t.Fatalf("expected passthrough error, got %v", err)
	}

	var resent int
	r := Retry{
		Transceiver: &Transceiver{
			Conn:    &stubConn{},
			Encoder: codec.NewEncoder(),
			Synchronizer: stubSynchronizer{
				writeToFn: func(io.Writer) (int64, error) {
					resent++
					return 0, nil
				},
			},
		},
		Times: 1,
	}
	if err := r.Send(ErrDeadlineExceeded); err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}
	if resent != 1 {
		t.Fatalf("expected resend to be called once, got %d", resent)
	}
}

func TestRetryRecv(t *testing.T) {
	packet := encodePacket(t, newPayloadEvent("recv"))
	conn := &readSequenceConn{reads: []readResult{{data: packet}}}
	tr := &Transceiver{
		Conn:         conn,
		Decoder:      codec.NewDecoder(gtp.DefaultMsgCreator()),
		Synchronizer: NewUnsequencedSynchronizer(),
	}

	e, err := (Retry{Transceiver: tr, Times: 1, Ctx: context.Background()}).Recv(IEvent{}, ErrDeadlineExceeded)
	if err != nil {
		t.Fatalf("expected retry recv success, got %v", err)
	}
	if string(AssertEvent[*gtp.MsgPayload](e).Msg.Data) != "recv" {
		t.Fatalf("unexpected event after retry: %+v", e)
	}

	e, err = (Retry{}).Recv(IEvent{}, retryErrTest)
	if !errors.Is(err, retryErrTest) {
		t.Fatalf("expected passthrough error, got %v", err)
	}
}
