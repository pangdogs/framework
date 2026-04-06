package variant

import (
	"errors"
	"reflect"
	"testing"

	"git.golaxy.org/core/utils/generic"
)

func TestCastVariant(t *testing.T) {
	var anyString any = "wrapped"
	arrInput := []any{"x", int32(3), true, nil}
	mapInput := map[string]any{"name": "demo", "count": int32(2)}
	sliceMapInput := generic.SliceMap[string, any]{{K: "name", V: "demo"}}
	unorderedInput := generic.UnorderedSliceMap[string, any]{{K: "name", V: "demo"}}

	cases := []struct {
		name   string
		input  any
		wantID TypeId
	}{
		{name: "int", input: int(7), wantID: TypeId_Int},
		{name: "int pointer", input: ptr(int(7)), wantID: TypeId_Int},
		{name: "uint8", input: uint8(8), wantID: TypeId_Uint8},
		{name: "string", input: "hello", wantID: TypeId_String},
		{name: "bytes", input: []byte("bytes"), wantID: TypeId_Bytes},
		{name: "bool", input: true, wantID: TypeId_Bool},
		{name: "nil", input: nil, wantID: TypeId_Null},
		{name: "wrapped any", input: &anyString, wantID: TypeId_String},
		{name: "array", input: arrInput, wantID: TypeId_Array},
		{name: "map", input: mapInput, wantID: TypeId_Map},
		{name: "slice map", input: sliceMapInput, wantID: TypeId_Map},
		{name: "unordered slice map", input: unorderedInput, wantID: TypeId_Map},
		{name: "error", input: errors.New("boom"), wantID: TypeId_Error},
		{name: "reflect value", input: reflect.ValueOf(int32(5)), wantID: TypeId_Int32},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := CastVariant(tc.input)
			if err != nil {
				t.Fatalf("CastVariant failed: %v", err)
			}
			if !v.IsValid() {
				t.Fatal("expected valid variant")
			}
			if v.TypeId != tc.wantID {
				t.Fatalf("unexpected type id: got %d want %d", v.TypeId, tc.wantID)
			}
		})
	}

	if _, err := CastVariant(make(chan int)); !errors.Is(err, ErrInvalidCast) {
		t.Fatalf("expected ErrInvalidCast, got %v", err)
	}
}

func TestCastSerializedVariant(t *testing.T) {
	arrInput := []any{"x", int32(3), true, nil}
	mapInput := map[string]any{"name": "demo", "count": int32(2)}

	cases := []struct {
		name   string
		input  any
		wantID TypeId
	}{
		{name: "string", input: "hello", wantID: TypeId_String},
		{name: "array", input: arrInput, wantID: TypeId_Array},
		{name: "map", input: mapInput, wantID: TypeId_Map},
		{name: "error", input: errors.New("boom"), wantID: TypeId_Error},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := CastSerializedVariant(tc.input)
			if err != nil {
				t.Fatalf("CastSerializedVariant failed: %v", err)
			}
			defer v.Release()

			if !v.IsValid() {
				t.Fatal("expected valid variant")
			}
			if v.TypeId != tc.wantID {
				t.Fatalf("unexpected type id: got %d want %d", v.TypeId, tc.wantID)
			}
			if _, ok := v.Value.(*SerializedValue); !ok {
				t.Fatalf("expected SerializedValue, got %T", v.Value)
			}
		})
	}
}

func TestVariantConvert(t *testing.T) {
	mapVariant, err := CastVariant(map[string]any{"name": "demo", "count": int32(2)})
	if err != nil {
		t.Fatalf("CastVariant map failed: %v", err)
	}

	mapRV, err := mapVariant.Convert(reflect.TypeFor[map[string]any]())
	if err != nil {
		t.Fatalf("Convert map failed: %v", err)
	}
	gotMap := mapRV.Interface().(map[string]any)
	if gotMap["name"] != "demo" || gotMap["count"] != int32(2) {
		t.Fatalf("unexpected converted map: %#v", gotMap)
	}

	arrVariant, err := CastVariant([]any{"x", int32(3), true})
	if err != nil {
		t.Fatalf("CastVariant array failed: %v", err)
	}

	arrRV, err := arrVariant.Convert(reflect.TypeFor[[]any]())
	if err != nil {
		t.Fatalf("Convert array failed: %v", err)
	}
	gotArr := arrRV.Interface().([]any)
	if len(gotArr) != 3 || gotArr[0] != "x" || gotArr[1] != int32(3) || gotArr[2] != true {
		t.Fatalf("unexpected converted array: %#v", gotArr)
	}

	intVariant, err := CastVariant(int32(9))
	if err != nil {
		t.Fatalf("CastVariant int failed: %v", err)
	}
	intRV, err := intVariant.Convert(reflect.TypeFor[int32]())
	if err != nil {
		t.Fatalf("Convert int failed: %v", err)
	}
	if got := intRV.Interface().(int32); got != 9 {
		t.Fatalf("unexpected converted int: %d", got)
	}

	nullVariant, err := CastVariant(nil)
	if err != nil {
		t.Fatalf("CastVariant nil failed: %v", err)
	}
	nilMapRV, err := nullVariant.Convert(reflect.TypeFor[map[string]any]())
	if err != nil {
		t.Fatalf("Convert nil map failed: %v", err)
	}
	if !nilMapRV.IsNil() {
		t.Fatal("expected nil map from null variant")
	}

	if _, err := intVariant.Convert(reflect.TypeFor[complex64]()); !errors.Is(err, ErrInvalidCast) {
		t.Fatalf("expected ErrInvalidCast, got %v", err)
	}
}

func ptr[T any](v T) *T {
	return &v
}
