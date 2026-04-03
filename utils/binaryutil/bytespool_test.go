package binaryutil

import (
	"bytes"
	"testing"
)

func TestNewBytesAndPayload(t *testing.T) {
	bs := NewBytes(false, -1)

	if got := len(bs.Payload()); got != 0 {
		t.Fatalf("unexpected payload length: got %d want 0", got)
	}
	if bs.Recyclable() {
		t.Fatal("expected non-recyclable bytes")
	}
}

func TestCloneBytesCopiesInput(t *testing.T) {
	src := []byte("hello")
	bs := CloneBytes(false, src)

	src[0] = 'H'

	if got := string(bs.Payload()); got != "hello" {
		t.Fatalf("unexpected cloned payload: got %q want %q", got, "hello")
	}
}

func TestRefBytesReferencesInput(t *testing.T) {
	src := []byte("hello")
	bs := RefBytes(src)

	src[0] = 'H'

	if got := string(bs.Payload()); got != "Hello" {
		t.Fatalf("unexpected referenced payload: got %q want %q", got, "Hello")
	}
	if bs.Recyclable() {
		t.Fatal("expected referenced bytes to be non-recyclable")
	}
}

func TestBytesSameRefAndSlice(t *testing.T) {
	root := NewBytes(false, 6)
	copy(root.Payload(), []byte("abcdef"))

	slice := root.Slice(1, 4)
	clone := CloneBytes(false, slice.Payload())

	if !root.SameRef(slice) {
		t.Fatal("expected slice to keep same backing reference")
	}
	if slice.SameRef(clone) {
		t.Fatal("expected clone to have different backing reference")
	}
	if got := string(slice.Payload()); got != "bcd" {
		t.Fatalf("unexpected slice payload: got %q want %q", got, "bcd")
	}
}

func TestBytesSlicePanicsOnInvalidRange(t *testing.T) {
	root := NewBytes(false, 4)

	tests := []struct {
		name string
		fn   func()
	}{
		{name: "negative", fn: func() { _ = root.Slice(-1, 1) }},
		{name: "reversed", fn: func() { _ = root.Slice(3, 2) }},
		{name: "out_of_range", fn: func() { _ = root.Slice(0, 5) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			tc.fn()
		})
	}
}

func TestReleaseAcceptsRecyclableAndNonRecyclable(t *testing.T) {
	recyclable := NewBytes(true, 8)
	copy(recyclable.Payload(), bytes.Repeat([]byte{'x'}, 8))
	recyclable.Release()

	nonRecyclable := NewBytes(false, 8)
	nonRecyclable.Release()
}
