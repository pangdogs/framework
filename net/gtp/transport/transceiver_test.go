package transport

import (
	"bytes"
	"context"
	"errors"
	"net"
	"testing"

	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/codec"
)

func TestTransceiverSendRecvRoundTrip(t *testing.T) {
	client, server := newUnsequencedPipeTransceivers(t)

	want := Event[*gtp.MsgPayload]{
		Flags: gtp.Flags_None().Setd(gtp.Flag_Compressed, true),
		Msg:   &gtp.MsgPayload{Data: []byte("hello")},
	}.Interface()

	done := make(chan error, 1)
	go func() {
		done <- client.Send(want)
	}()

	got := recvWithTimeout(t, server)
	if err := <-done; err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if !bytes.Equal(AssertEvent[*gtp.MsgPayload](got).Msg.Data, []byte("hello")) {
		t.Fatalf("unexpected payload: %q", AssertEvent[*gtp.MsgPayload](got).Msg.Data)
	}
}

func TestTransceiverSendRst(t *testing.T) {
	client, server := newUnsequencedPipeTransceivers(t)

	done := make(chan error, 1)
	go func() {
		done <- client.SendRst(errors.New("boom"))
	}()

	got := recvWithTimeout(t, server)
	if err := <-done; err != nil {
		t.Fatalf("SendRst failed: %v", err)
	}
	rst := AssertEvent[*gtp.MsgRst](got)
	if rst.Msg.Message != "boom" {
		t.Fatalf("unexpected rst message: %q", rst.Msg.Message)
	}
}

func TestTransceiverSendRecvResendErrors(t *testing.T) {
	t.Run("send nil deps", func(t *testing.T) {
		tr := &Transceiver{}
		if err := tr.Send(newPayloadEvent("x")); err == nil {
			t.Fatal("expected send error")
		}
	})

	t.Run("recv nil deps", func(t *testing.T) {
		tr := &Transceiver{}
		if _, err := tr.Recv(context.Background()); err == nil {
			t.Fatal("expected recv error")
		}
	})

	t.Run("resend nil deps", func(t *testing.T) {
		tr := &Transceiver{}
		if err := tr.Resend(); err == nil {
			t.Fatal("expected resend error")
		}
	})

	t.Run("recv canceled", func(t *testing.T) {
		clientConn, serverConn := netPipeNoCleanup(t)
		defer clientConn.Close()
		defer serverConn.Close()

		tr := &Transceiver{
			Conn:         clientConn,
			Decoder:      codec.NewDecoder(gtp.DefaultMsgCreator()),
			Synchronizer: NewUnsequencedSynchronizer(),
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := tr.Recv(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context canceled, got %v", err)
		}
	})
}

func TestTransceiverResendAfterDeadline(t *testing.T) {
	var writes int
	var sent bytes.Buffer

	conn := &stubConn{
		writeFn: func(p []byte) (int, error) {
			writes++
			if writes == 1 {
				return 0, ErrDeadlineExceeded
			}
			return sent.Write(p)
		},
	}
	tr := &Transceiver{
		Conn:         conn,
		Encoder:      codec.NewEncoder(),
		Decoder:      codec.NewDecoder(gtp.DefaultMsgCreator()),
		Synchronizer: NewSequencedSynchronizer(1, 1, 4096),
	}

	err := tr.Send(newPayloadEvent("retry"))
	if !errors.Is(err, ErrDeadlineExceeded) {
		t.Fatalf("expected deadline error, got %v", err)
	}

	if err := tr.Resend(); err != nil {
		t.Fatalf("Resend failed: %v", err)
	}
	if sent.Len() == 0 {
		t.Fatal("expected resent bytes")
	}
}

func TestTransceiverMigrateAndDispose(t *testing.T) {
	var synchronized uint32
	var disposed bool

	oldConn, newConn := &stubConn{}, &stubConn{}
	tr := &Transceiver{
		Conn: oldConn,
		Synchronizer: stubSynchronizer{
			syncFn: func(seq uint32) error {
				synchronized = seq
				return nil
			},
			sendSeqFn: func() uint32 { return 3 },
			recvSeqFn: func() uint32 { return 4 },
			disposeFn: func() { disposed = true },
		},
	}
	tr.buffer.WriteString("cached")

	sendReq, recvReq, err := tr.Migrate(newConn, 9)
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
	if synchronized != 9 || sendReq != 3 || recvReq != 4 {
		t.Fatalf("unexpected migrate result: sync=%d send=%d recv=%d", synchronized, sendReq, recvReq)
	}
	if tr.buffer.Len() != 0 {
		t.Fatal("expected buffer reset after migrate")
	}

	tr.Dispose()
	if !disposed {
		t.Fatal("expected synchronizer dispose")
	}
}

func netPipeNoCleanup(t *testing.T) (net.Conn, net.Conn) {
	t.Helper()
	return net.Pipe()
}
