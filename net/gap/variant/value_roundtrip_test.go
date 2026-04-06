package variant

import "testing"

func TestBuiltinValueRoundTrip(t *testing.T) {
	for _, tc := range sampleValues(t) {
		t.Run(tc.name, func(t *testing.T) {
			payload := encodeReadable(t, tc.value, tc.value.Size())
			decoded := decodeValue(t, tc.value.TypeId(), payload)
			decodedPayload := encodeReadable(t, decoded, decoded.Size())

			if decoded.TypeId() != tc.value.TypeId() {
				t.Fatalf("unexpected type id: got %d want %d", decoded.TypeId(), tc.value.TypeId())
			}
			compareBytes(t, decodedPayload, payload)
		})
	}
}

func TestVariantRoundTrip(t *testing.T) {
	for _, tc := range sampleValues(t) {
		t.Run(tc.name, func(t *testing.T) {
			v, err := NewVariant(tc.value)
			if err != nil {
				t.Fatalf("NewVariant failed: %v", err)
			}

			wire := encodeVariant(t, v)
			got := decodeVariant(t, wire)
			gotWire := encodeVariant(t, got)

			if got.TypeId != tc.value.TypeId() {
				t.Fatalf("unexpected decoded type id: got %d want %d", got.TypeId, tc.value.TypeId())
			}
			compareBytes(t, gotWire, wire)
		})
	}
}

func TestSerializedVariantRoundTrip(t *testing.T) {
	for _, tc := range sampleValues(t) {
		t.Run(tc.name, func(t *testing.T) {
			v, err := NewSerializedVariant(tc.value)
			if err != nil {
				t.Fatalf("NewSerializedVariant failed: %v", err)
			}
			defer v.Release()

			wire := encodeVariant(t, v)
			got := decodeVariant(t, wire)
			gotWire := encodeVariant(t, got)

			if got.TypeId != tc.value.TypeId() {
				t.Fatalf("unexpected decoded type id: got %d want %d", got.TypeId, tc.value.TypeId())
			}
			compareBytes(t, gotWire, wire)
		})
	}
}

func TestNewVariantErrors(t *testing.T) {
	if _, err := NewVariant(nil); err == nil {
		t.Fatal("expected NewVariant error for nil")
	}
	if _, err := NewSerializedVariant(nil); err == nil {
		t.Fatal("expected NewSerializedVariant error for nil")
	}
	if _, err := NewSerializedValue(nil); err == nil {
		t.Fatal("expected NewSerializedValue error for nil")
	}
}
