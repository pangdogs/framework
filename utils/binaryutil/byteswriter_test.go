package binaryutil

import (
	"bytes"
	"testing"
)

func TestBytesWriterWrite(t *testing.T) {
	writer := NewBytesWriter(make([]byte, 5))

	n, err := writer.Write([]byte("abc"))
	if err != nil {
		t.Fatalf("first write failed: %v", err)
	}
	if n != 3 {
		t.Fatalf("unexpected bytes written: got %d want 3", n)
	}

	n, err = writer.Write([]byte("de"))
	if err != nil {
		t.Fatalf("second write failed: %v", err)
	}
	if n != 2 {
		t.Fatalf("unexpected bytes written: got %d want 2", n)
	}
	if writer.N != 5 {
		t.Fatalf("unexpected write offset: got %d want 5", writer.N)
	}
	if got := writer.Bytes; !bytes.Equal(got, []byte("abcde")) {
		t.Fatalf("unexpected buffer: got %q want %q", got, "abcde")
	}
}

func TestBytesWriterWriteLimitReached(t *testing.T) {
	writer := NewBytesWriter(make([]byte, 4))

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
		t.Fatalf("unexpected write offset: got %d want 2", writer.N)
	}
	if got := writer.Bytes; !bytes.Equal(got, []byte{'a', 'b', 0, 0}) {
		t.Fatalf("unexpected buffer contents: %v", got)
	}
}

func TestBytesWriterWriteAfterFull(t *testing.T) {
	writer := NewBytesWriter([]byte("ab"))
	writer.N = len(writer.Bytes)

	n, err := writer.Write([]byte("x"))
	if err != ErrLimitReached {
		t.Fatalf("unexpected error: got %v want %v", err, ErrLimitReached)
	}
	if n != 0 {
		t.Fatalf("unexpected bytes written: got %d want 0", n)
	}
}
