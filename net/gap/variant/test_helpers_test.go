package variant

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"
)

func sampleValues(t *testing.T) []struct {
	name  string
	value Value
} {
	t.Helper()

	arr, err := NewArray([]any{int64(7), "nested", true, nil})
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	m, err := NewMapFromGoMap(map[string]any{
		"count": int32(9),
		"label": "value",
		"ok":    true,
	})
	if err != nil {
		t.Fatalf("NewMapFromGoMap failed: %v", err)
	}

	return []struct {
		name  string
		value Value
	}{
		{name: "Int", value: ptrValue(Int(-11))},
		{name: "Int8", value: ptrValue(Int8(-12))},
		{name: "Int16", value: ptrValue(Int16(-1234))},
		{name: "Int32", value: ptrValue(Int32(-123456))},
		{name: "Int64", value: ptrValue(Int64(-123456789))},
		{name: "Uint", value: ptrValue(Uint(11))},
		{name: "Uint8", value: ptrValue(Uint8(12))},
		{name: "Uint16", value: ptrValue(Uint16(1234))},
		{name: "Uint32", value: ptrValue(Uint32(123456))},
		{name: "Uint64", value: ptrValue(Uint64(123456789))},
		{name: "Float", value: ptrValue(Float(12.5))},
		{name: "Double", value: ptrValue(Double(123.75))},
		{name: "Byte", value: ptrValue(Byte(0xab))},
		{name: "Bool", value: ptrValue(Bool(true))},
		{name: "Bytes", value: ptrValue(Bytes([]byte("payload")))},
		{name: "String", value: ptrValue(String("hello"))},
		{name: "Null", value: &Null{}},
		{name: "Array", value: &arr},
		{name: "Map", value: &m},
		{name: "Error", value: Errorf(7, "broken-%s", "state")},
		{name: "CallChain", value: &CallChain{
			{Svc: "svc-a", Addr: "127.0.0.1", Timestamp: time.UnixMilli(123456).UTC(), Transit: false},
			{Svc: "svc-b", Addr: "127.0.0.2", Timestamp: time.UnixMilli(223456).UTC(), Transit: true},
		}},
	}
}

func ptrValue[T any](v T) *T {
	return &v
}

func encodeReadable(t *testing.T, r io.Reader, size int) []byte {
	t.Helper()

	buf := make([]byte, size)
	n, err := r.Read(buf)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("Read failed: n=%d err=%v", n, err)
	}
	if n != size {
		t.Fatalf("Read size mismatch: got %d want %d", n, size)
	}

	return buf
}

func decodeValue(t *testing.T, typeID TypeId, data []byte) Value {
	t.Helper()

	v, err := typeID.New()
	if err != nil {
		t.Fatalf("TypeId.New failed: %v", err)
	}

	n, err := v.Write(data)
	if err != nil {
		t.Fatalf("Write failed: n=%d err=%v", n, err)
	}
	if n != len(data) {
		t.Fatalf("Write size mismatch: got %d want %d", n, len(data))
	}

	return v
}

func encodeVariant(t *testing.T, v Variant) []byte {
	t.Helper()
	return encodeReadable(t, v, v.Size())
}

func encodeSerializedVariant(t *testing.T, v SerializedVariant) []byte {
	t.Helper()
	return encodeVariant(t, v.Ref())
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

func requirePanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	fn()
}

func compareBytes(t *testing.T, got, want []byte) {
	t.Helper()
	if !bytes.Equal(got, want) {
		t.Fatalf("bytes mismatch: got %v want %v", got, want)
	}
}
