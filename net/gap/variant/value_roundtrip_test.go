package variant

import (
	"testing"
	"time"
)

func TestVariantRoundTripBuiltins(t *testing.T) {
	now := time.UnixMilli(1710000000123).Local()

	tests := []struct {
		name   string
		input  any
		typeID TypeId
	}{
		{name: "int", input: int(-1), typeID: TypeId_Int},
		{name: "int8", input: int8(-2), typeID: TypeId_Int8},
		{name: "int16", input: int16(-3), typeID: TypeId_Int16},
		{name: "int32", input: int32(-4), typeID: TypeId_Int32},
		{name: "int64", input: int64(-5), typeID: TypeId_Int64},
		{name: "uint", input: uint(1), typeID: TypeId_Uint},
		{name: "uint8", input: uint8(2), typeID: TypeId_Uint8},
		{name: "uint16", input: uint16(3), typeID: TypeId_Uint16},
		{name: "uint32", input: uint32(4), typeID: TypeId_Uint32},
		{name: "uint64", input: uint64(5), typeID: TypeId_Uint64},
		{name: "float32", input: float32(1.5), typeID: TypeId_Float},
		{name: "float64", input: float64(2.5), typeID: TypeId_Double},
		{name: "bool", input: true, typeID: TypeId_Bool},
		{name: "bytes", input: []byte("payload"), typeID: TypeId_Bytes},
		{name: "string", input: "hello", typeID: TypeId_String},
		{name: "null", input: nil, typeID: TypeId_Null},
		{name: "error", input: Errorln(7, "boom"), typeID: TypeId_Error},
		{name: "callchain", input: CallChain{{Svc: "svc", Addr: "addr", Timestamp: now, Transit: true}}, typeID: TypeId_CallChain},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v, err := CastVariant(tc.input)
			if err != nil {
				t.Fatalf("CastVariant failed: %v", err)
			}
			if v.TypeId != tc.typeID {
				t.Fatalf("type id mismatch: got %d want %d", v.TypeId, tc.typeID)
			}

			got := assertWireRoundTrip(t, v)
			if got.TypeId != tc.typeID {
				t.Fatalf("decoded type id mismatch: got %d want %d", got.TypeId, tc.typeID)
			}
		})
	}
}

func TestVariantRoundTripArrayAndMap(t *testing.T) {
	arr, err := NewArray([]any{"x", int32(3), true, nil})
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	arrVariant, err := NewVariant(arr)
	if err != nil {
		t.Fatalf("NewVariant array failed: %v", err)
	}
	gotArrVariant := assertWireRoundTrip(t, arrVariant)
	gotArr, ok := gotArrVariant.Value.(*Array)
	if !ok {
		t.Fatalf("decoded array value type = %T, want *Array", gotArrVariant.Value)
	}
	if len(gotArr.Items) != 4 {
		t.Fatalf("decoded array length = %d, want 4", len(gotArr.Items))
	}

	m, err := NewMapFromGoMap(map[string]any{"name": "demo", "count": int32(2)})
	if err != nil {
		t.Fatalf("NewMapFromGoMap failed: %v", err)
	}

	mapVariant, err := NewVariant(m)
	if err != nil {
		t.Fatalf("NewVariant map failed: %v", err)
	}
	gotMapVariant := assertWireRoundTrip(t, mapVariant)
	gotMap, ok := gotMapVariant.Value.(*Map)
	if !ok {
		t.Fatalf("decoded map value type = %T, want *Map", gotMapVariant.Value)
	}
	if len(gotMap.Entries) != 2 {
		t.Fatalf("decoded map length = %d, want 2", len(gotMap.Entries))
	}
}
