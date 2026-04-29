package transport

import (
	"bytes"
	"errors"
	"io"
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

	s.ack(2)
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

func TestSequencedSynchronizerAckClearsConfirmedFrames(t *testing.T) {
	t.Run("ack next expected keeps first unconfirmed frame", func(t *testing.T) {
		s := NewSequencedSynchronizer(1, 1, 4096).(*SequencedSynchronizer)

		packets := [][]byte{
			encodePacket(t, newPayloadEvent("one")),
			encodePacket(t, newPayloadEvent("two")),
			encodePacket(t, newPayloadEvent("three")),
		}
		for _, packet := range packets {
			if _, err := s.Write(packet); err != nil {
				t.Fatalf("Write failed: %v", err)
			}
		}

		s.Ack(2)

		if _, err := s.Write(encodePacket(t, newPayloadEvent("four"))); err != nil {
			t.Fatalf("Write after ack failed: %v", err)
		}

		if got, want := s.queue.Length(), 3; got != want {
			t.Fatalf("unexpected queue length after ack cleanup: got %d want %d", got, want)
		}

		gotSeqs := []uint32{
			s.queue.Index(0).seq,
			s.queue.Index(1).seq,
			s.queue.Index(2).seq,
		}
		wantSeqs := []uint32{2, 3, 4}
		for i := range wantSeqs {
			if gotSeqs[i] != wantSeqs[i] {
				t.Fatalf("unexpected queue seqs after ack cleanup: got %v want %v", gotSeqs, wantSeqs)
			}
		}
	})

	t.Run("ack beyond current tail still clears older confirmed frame", func(t *testing.T) {
		s := NewSequencedSynchronizer(1, 1, 4096).(*SequencedSynchronizer)

		packet := encodePacket(t, newPayloadEvent("one"))
		if _, err := s.Write(packet); err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		s.Ack(2)

		if _, err := s.Write(encodePacket(t, newPayloadEvent("two"))); err != nil {
			t.Fatalf("Write after ack failed: %v", err)
		}

		if got, want := s.queue.Length(), 1; got != want {
			t.Fatalf("unexpected queue length after tail ack cleanup: got %d want %d", got, want)
		}
		if got, want := s.queue.Index(0).seq, uint32(2); got != want {
			t.Fatalf("unexpected remaining seq after tail ack cleanup: got %d want %d", got, want)
		}
	})
}

func TestSequencedSynchronizerAckShiftsSentCursor(t *testing.T) {
	s := NewSequencedSynchronizer(1, 1, 4096).(*SequencedSynchronizer)

	for _, payload := range []string{"one", "two", "three"} {
		if _, err := s.Write(encodePacket(t, newPayloadEvent(payload))); err != nil {
			t.Fatalf("Write(%q) failed: %v", payload, err)
		}
	}

	var sent bytes.Buffer
	if _, err := s.WriteTo(&sent); err != nil {
		t.Fatalf("initial WriteTo failed: %v", err)
	}
	if got, want := s.sent, 3; got != want {
		t.Fatalf("unexpected sent cursor before ack: got %d want %d", got, want)
	}

	s.Ack(2)
	if _, err := s.Write(encodePacket(t, newPayloadEvent("four"))); err != nil {
		t.Fatalf("Write after ack failed: %v", err)
	}

	if got, want := s.sent, 2; got != want {
		t.Fatalf("unexpected sent cursor after ack cleanup: got %d want %d", got, want)
	}

	var resend bytes.Buffer
	if _, err := s.WriteTo(&resend); err != nil {
		t.Fatalf("follow-up WriteTo failed: %v", err)
	}

	heads := decodePacketHeads(t, resend.Bytes())
	if got, want := len(heads), 1; got != want {
		t.Fatalf("unexpected resent packet count: got %d want %d", got, want)
	}
	if got, want := heads[0].Seq, uint32(4); got != want {
		t.Fatalf("unexpected resent seq after ack cleanup: got %d want %d", got, want)
	}
}

func TestSequencedSynchronizerWriteToShortWritePanics(t *testing.T) {
	s := NewSequencedSynchronizer(1, 1, 4096).(*SequencedSynchronizer)

	packet := encodePacket(t, newPayloadEvent("partial-send"))
	if _, err := s.Write(packet); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	w := &shortWriteRecorder{limit: len(packet) / 2}
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
		if got, want := len(w.buf.Bytes()), len(packet)/2; got != want {
			t.Fatalf("unexpected first write size: got %d want %d", got, want)
		}
		if got, want := s.queue.Index(0).offset, len(packet)/2; got != want {
			t.Fatalf("unexpected frame offset after short write: got %d want %d", got, want)
		}
		if s.sent != 0 {
			t.Fatalf("expected sent cursor to remain at 0, got %d", s.sent)
		}
	}()

	if _, err := s.WriteTo(w); err != nil {
		t.Fatalf("unexpected WriteTo error before panic: %v", err)
	}
}

func TestSequencedSynchronizerWraparound(t *testing.T) {
	maxSeq := ^uint32(0)

	t.Run("write and ack wrap around sequence space", func(t *testing.T) {
		s := NewSequencedSynchronizer(maxSeq-1, 1, 4096).(*SequencedSynchronizer)

		for _, payload := range []string{"one", "two", "three"} {
			if _, err := s.Write(encodePacket(t, newPayloadEvent(payload))); err != nil {
				t.Fatalf("Write(%q) failed: %v", payload, err)
			}
		}

		if got, want := s.SendSeq(), uint32(1); got != want {
			t.Fatalf("unexpected wrapped send seq: got %d want %d", got, want)
		}

		gotSeqs := []uint32{
			s.queue.Index(0).seq,
			s.queue.Index(1).seq,
			s.queue.Index(2).seq,
		}
		wantSeqs := []uint32{maxSeq - 1, maxSeq, 0}
		for i := range wantSeqs {
			if gotSeqs[i] != wantSeqs[i] {
				t.Fatalf("unexpected wrapped queue seqs: got %v want %v", gotSeqs, wantSeqs)
			}
		}

		s.Ack(0)
		if _, err := s.Write(encodePacket(t, newPayloadEvent("four"))); err != nil {
			t.Fatalf("Write after wrapped ack failed: %v", err)
		}

		if got, want := s.queue.Length(), 2; got != want {
			t.Fatalf("unexpected queue length after wrapped ack cleanup: got %d want %d", got, want)
		}
		if got, want := s.queue.Index(0).seq, uint32(0); got != want {
			t.Fatalf("unexpected remaining wrapped seq: got %d want %d", got, want)
		}
		if got, want := s.queue.Index(1).seq, uint32(1); got != want {
			t.Fatalf("unexpected appended wrapped seq: got %d want %d", got, want)
		}
	})

	t.Run("validate handles wrapped recv sequence", func(t *testing.T) {
		s := NewSequencedSynchronizer(1, maxSeq, 4096).(*SequencedSynchronizer)

		if err := s.Validate(gtp.MsgHead{Seq: maxSeq}, nil); err != nil {
			t.Fatalf("expected wrapped current seq to validate, got %v", err)
		}

		s.Ack(0)
		if got, want := s.RecvSeq(), uint32(0); got != want {
			t.Fatalf("unexpected wrapped recv seq: got %d want %d", got, want)
		}

		if err := s.Validate(gtp.MsgHead{Seq: 0}, nil); err != nil {
			t.Fatalf("expected wrapped next seq to validate, got %v", err)
		}
		if err := s.Validate(gtp.MsgHead{Seq: maxSeq}, nil); !errors.Is(err, ErrDiscardSeq) {
			t.Fatalf("expected wrapped previous seq discard, got %v", err)
		}
		if err := s.Validate(gtp.MsgHead{Seq: 1}, nil); !errors.Is(err, ErrUnexpectedSeq) {
			t.Fatalf("expected wrapped future seq unexpected, got %v", err)
		}
	})

	t.Run("synchronize replays from wrapped sequence", func(t *testing.T) {
		s := NewSequencedSynchronizer(maxSeq-1, 1, 4096).(*SequencedSynchronizer)

		for _, payload := range []string{"one", "two", "three"} {
			if _, err := s.Write(encodePacket(t, newPayloadEvent(payload))); err != nil {
				t.Fatalf("Write(%q) failed: %v", payload, err)
			}
		}

		if err := s.Synchronize(maxSeq); err != nil {
			t.Fatalf("Synchronize failed: %v", err)
		}

		var out bytes.Buffer
		if _, err := s.WriteTo(&out); err != nil {
			t.Fatalf("WriteTo after wrapped synchronize failed: %v", err)
		}

		heads := decodePacketHeads(t, out.Bytes())
		gotSeqs := make([]uint32, 0, len(heads))
		for _, head := range heads {
			gotSeqs = append(gotSeqs, head.Seq)
		}
		wantSeqs := []uint32{maxSeq, 0}
		if got, want := len(gotSeqs), len(wantSeqs); got != want {
			t.Fatalf("unexpected wrapped replay count: got %d want %d", got, want)
		}
		for i := range wantSeqs {
			if gotSeqs[i] != wantSeqs[i] {
				t.Fatalf("unexpected wrapped replay seqs: got %v want %v", gotSeqs, wantSeqs)
			}
		}
	})
}

func TestSequencedSynchronizerSynchronizeAffectsNextWriteWindow(t *testing.T) {
	t.Run("next write clears prefix before replay start", func(t *testing.T) {
		s := NewSequencedSynchronizer(1, 1, 4096).(*SequencedSynchronizer)

		for _, payload := range []string{"one", "two", "three"} {
			if _, err := s.Write(encodePacket(t, newPayloadEvent(payload))); err != nil {
				t.Fatalf("Write(%q) failed: %v", payload, err)
			}
		}

		if err := s.Synchronize(2); err != nil {
			t.Fatalf("Synchronize failed: %v", err)
		}
		if _, err := s.Write(encodePacket(t, newPayloadEvent("four"))); err != nil {
			t.Fatalf("Write after synchronize failed: %v", err)
		}

		gotSeqs := []uint32{
			s.queue.Index(0).seq,
			s.queue.Index(1).seq,
			s.queue.Index(2).seq,
		}
		wantSeqs := []uint32{2, 3, 4}
		for i := range wantSeqs {
			if gotSeqs[i] != wantSeqs[i] {
				t.Fatalf("unexpected queue seqs after synchronize/write: got %v want %v", gotSeqs, wantSeqs)
			}
		}
	})

	t.Run("conservative replay keeps last frame when remote recv seq equals send seq", func(t *testing.T) {
		s := NewSequencedSynchronizer(1, 1, 4096).(*SequencedSynchronizer)

		for _, payload := range []string{"one", "two"} {
			if _, err := s.Write(encodePacket(t, newPayloadEvent(payload))); err != nil {
				t.Fatalf("Write(%q) failed: %v", payload, err)
			}
		}

		if err := s.Synchronize(3); err != nil {
			t.Fatalf("Synchronize failed: %v", err)
		}
		if _, err := s.Write(encodePacket(t, newPayloadEvent("three"))); err != nil {
			t.Fatalf("Write after synchronize failed: %v", err)
		}

		if got, want := s.queue.Length(), 2; got != want {
			t.Fatalf("unexpected queue length after conservative replay write: got %d want %d", got, want)
		}
		if got, want := s.queue.Index(0).seq, uint32(2); got != want {
			t.Fatalf("unexpected replay start seq after conservative replay write: got %d want %d", got, want)
		}
		if got, want := s.queue.Index(1).seq, uint32(3); got != want {
			t.Fatalf("unexpected appended seq after conservative replay write: got %d want %d", got, want)
		}
	})
}

type shortWriteRecorder struct {
	buf   bytes.Buffer
	limit int
	calls int
}

func (w *shortWriteRecorder) Write(p []byte) (int, error) {
	w.calls++
	if w.calls == 1 && w.limit > 0 && w.limit < len(p) {
		return w.buf.Write(p[:w.limit])
	}
	return w.buf.Write(p)
}

var _ io.Writer = (*shortWriteRecorder)(nil)
