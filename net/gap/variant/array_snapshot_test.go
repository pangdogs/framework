package variant

import (
	"bytes"
	"errors"
	"testing"
)

func TestArraySnapshotReadSizeAndReadonly(t *testing.T) {
	arr, err := NewArray([]any{"x", int32(3), true})
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	snapshot, err := arr.Snapshot(false)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if !snapshot.IsSnapshot {
		t.Fatal("expected snapshot array")
	}
	if len(snapshot.Items) != 0 {
		t.Fatalf("snapshot items length = %d, want 0", len(snapshot.Items))
	}
	if snapshot.Size() != arr.Size() {
		t.Fatalf("snapshot size = %d want %d", snapshot.Size(), arr.Size())
	}

	arrWire := encodeValue(t, arr)
	snapshotWire := encodeValue(t, snapshot)
	if !bytes.Equal(snapshotWire, arrWire) {
		t.Fatalf("snapshot wire mismatch:\ngot  %v\nwant %v", snapshotWire, arrWire)
	}

	var decoded Array
	if n, err := decoded.Write(snapshotWire); err != nil || n != len(snapshotWire) {
		t.Fatalf("decode snapshot payload failed: n=%d err=%v", n, err)
	}
	if decoded.IsSnapshot {
		t.Fatal("decoded array should not be a snapshot")
	}
	if len(decoded.Items) != len(arr.Items) {
		t.Fatalf("decoded items length = %d want %d", len(decoded.Items), len(arr.Items))
	}

	if n, err := snapshot.Write(arrWire); !errors.Is(err, ErrSnapshotReadonly) || n != 0 {
		t.Fatalf("snapshot Write = (%d, %v), want (0, ErrSnapshotReadonly)", n, err)
	}
}

func TestSnapshotArrayDoesNotConvertAsNormalArray(t *testing.T) {
	arr, err := NewArray([]any{"x"})
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	snapshot, err := arr.Snapshot(false)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	v, err := NewVariant(snapshot)
	if err != nil {
		t.Fatalf("NewVariant failed: %v", err)
	}

	if _, err := v.ToNative(sliceAnyRT); !errors.Is(err, ErrInvalidCast) {
		t.Fatalf("snapshot ToNative error = %v, want ErrInvalidCast", err)
	}
}
