package binaryutil

import (
	"bytes"
	"errors"
	"testing"
)

func TestNewLimitWriterNormalizesNegativeLimit(t *testing.T) {
	writer := NewLimitWriter(&bytes.Buffer{}, -1)
	if writer.Limit != 0 {
		t.Fatalf("unexpected limit: got %d want 0", writer.Limit)
	}
}

func TestLimitWriterWrite(t *testing.T) {
	var out bytes.Buffer
	writer := NewLimitWriter(&out, 5)

	n, err := writer.Write([]byte("abc"))
	if err != nil {
		t.Fatalf("first write failed: %v", err)
	}
	if n != 3 {
		t.Fatalf("unexpected first write length: got %d want 3", n)
	}

	n, err = writer.Write([]byte("de"))
	if err != nil {
		t.Fatalf("second write failed: %v", err)
	}
	if n != 2 {
		t.Fatalf("unexpected second write length: got %d want 2", n)
	}
	if writer.N != 5 {
		t.Fatalf("unexpected bytes written count: got %d want 5", writer.N)
	}
	if got := out.String(); got != "abcde" {
		t.Fatalf("unexpected output: got %q want %q", got, "abcde")
	}
}

func TestLimitWriterRejectsOverflowBeforeWrite(t *testing.T) {
	var out bytes.Buffer
	writer := NewLimitWriter(&out, 4)

	if _, err := writer.Write([]byte("ab")); err != nil {
		t.Fatalf("seed write failed: %v", err)
	}

	n, err := writer.Write([]byte("cde"))
	if err != ErrLimitReached {
		t.Fatalf("unexpected error: got %v want %v", err, ErrLimitReached)
	}
	if n != 0 {
		t.Fatalf("unexpected bytes written: got %d want 0", n)
	}
	if writer.N != 2 {
		t.Fatalf("unexpected write count: got %d want 2", writer.N)
	}
	if got := out.String(); got != "ab" {
		t.Fatalf("unexpected output: got %q want %q", got, "ab")
	}
}

func TestLimitWriterPropagatesUnderlyingWriterResult(t *testing.T) {
	writer := NewLimitWriter(partialWriter{n: 2, err: errors.New("boom")}, 4)

	n, err := writer.Write([]byte("abcd"))
	if err == nil || err.Error() != "boom" {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("unexpected bytes written: got %d want 2", n)
	}
	if writer.N != 2 {
		t.Fatalf("unexpected write count: got %d want 2", writer.N)
	}
}

type partialWriter struct {
	n   int
	err error
}

func (w partialWriter) Write(p []byte) (int, error) {
	if w.n > len(p) {
		return len(p), w.err
	}
	return w.n, w.err
}
