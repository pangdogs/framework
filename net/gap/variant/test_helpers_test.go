package variant

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type readableSizer interface {
	io.Reader
	Size() int
}

func encodeValue(t *testing.T, v readableSizer) []byte {
	t.Helper()

	data := make([]byte, v.Size())
	n, err := v.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("Read failed: n=%d err=%v", n, err)
	}
	if n != len(data) {
		t.Fatalf("Read size mismatch: got %d want %d", n, len(data))
	}
	return data
}

func encodeVariant(t *testing.T, v Variant) []byte {
	t.Helper()
	return encodeValue(t, v)
}

func decodeVariant(t *testing.T, data []byte) Variant {
	t.Helper()

	var v Variant
	n, err := v.Write(data)
	if err != nil {
		t.Fatalf("Variant.Write failed: n=%d err=%v", n, err)
	}
	if n != len(data) {
		t.Fatalf("Variant.Write size mismatch: got %d want %d", n, len(data))
	}
	if !v.IsValid() {
		t.Fatal("decoded variant is invalid")
	}
	return v
}

func assertWireRoundTrip(t *testing.T, v Variant) Variant {
	t.Helper()

	wire := encodeVariant(t, v)
	got := decodeVariant(t, wire)
	gotWire := encodeVariant(t, got)
	if !bytes.Equal(gotWire, wire) {
		t.Fatalf("roundtrip wire mismatch:\ngot  %v\nwant %v", gotWire, wire)
	}
	return got
}
