package transport

import (
	"bytes"
	"testing"

	"git.golaxy.org/framework/net/gtp"
)

func TestUnsequencedSynchronizerBasics(t *testing.T) {
	s := NewUnsequencedSynchronizer()
	if s == nil {
		t.Fatal("expected synchronizer")
	}

	if _, err := s.Write([]byte("abc")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if s.Cached() != 3 {
		t.Fatalf("unexpected cached size: %d", s.Cached())
	}

	var out bytes.Buffer
	n, err := s.WriteTo(&out)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if n != 3 || out.String() != "abc" {
		t.Fatalf("unexpected WriteTo result: n=%d out=%q", n, out.String())
	}

	if err := s.Validate(gtp.MsgHead{}, nil); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if err := s.Synchronize(1); err == nil {
		t.Fatal("expected synchronize error")
	}

	s.Dispose()
	if s.Cached() != 0 {
		t.Fatalf("expected empty cache after dispose, got %d", s.Cached())
	}
}

func TestUnsequencedSynchronizerWriteToNil(t *testing.T) {
	s := NewUnsequencedSynchronizer()
	if _, err := s.WriteTo(nil); err == nil {
		t.Fatal("expected nil writer error")
	}
}
