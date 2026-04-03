package transport

import (
	"bytes"
	"errors"
	"testing"

	"git.golaxy.org/framework/net/gtp"
)

func TestSequencedSynchronizerWriteValidateAckAndSynchronize(t *testing.T) {
	s := NewSequencedSynchronizer(1, 1, 4096).(*SequencedSynchronizer)

	packet := encodePacket(t, Event[*gtp.MsgPayload]{Msg: &gtp.MsgPayload{Data: []byte("abc")}}.Interface())
	if n, err := s.Write(packet); err != nil || n != len(packet) {
		t.Fatalf("Write failed: n=%d err=%v", n, err)
	}
	if s.SendSeq() != 2 {
		t.Fatalf("unexpected send seq: %d", s.SendSeq())
	}
	if s.Cached() != len(packet) {
		t.Fatalf("unexpected cached size: %d", s.Cached())
	}

	var out bytes.Buffer
	if _, err := s.WriteTo(&out); err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	var head gtp.MsgHead
	if _, err := head.Write(out.Bytes()); err != nil {
		t.Fatalf("head.Write failed: %v", err)
	}
	if head.Seq != 1 || head.Ack != 1 {
		t.Fatalf("unexpected frame head: %+v", head)
	}

	if err := s.Validate(head, nil); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	head.Seq = 2
	if err := s.Validate(head, nil); !errors.Is(err, ErrUnexpectedSeq) {
		t.Fatalf("expected ErrUnexpectedSeq, got %v", err)
	}
	head.Seq = 0
	if err := s.Validate(head, nil); !errors.Is(err, ErrDiscardSeq) {
		t.Fatalf("expected ErrDiscardSeq, got %v", err)
	}

	s.Ack(7)
	if s.RecvSeq() != 2 || s.AckSeq() != 7 {
		t.Fatalf("unexpected ack state: recv=%d ack=%d", s.RecvSeq(), s.AckSeq())
	}

	if err := s.Synchronize(1); err != nil {
		t.Fatalf("Synchronize failed: %v", err)
	}

	s.ack(1)
	if s.Cached() != 0 {
		t.Fatalf("expected empty cache after ack, got %d", s.Cached())
	}
}

func TestSequencedSynchronizerErrorsAndDispose(t *testing.T) {
	s := NewSequencedSynchronizer(1, 1, 1).(*SequencedSynchronizer)

	if _, err := s.Write([]byte("bad")); err == nil {
		t.Fatal("expected write error for invalid packet")
	}
	if _, err := s.WriteTo(nil); err == nil {
		t.Fatal("expected nil writer error")
	}
	if err := s.Synchronize(99); err == nil {
		t.Fatal("expected synchronize error")
	}

	s.Dispose()
	if s.Cached() != 0 || s.SendSeq() != 0 || s.RecvSeq() != 0 {
		t.Fatal("unexpected state after dispose")
	}
}
