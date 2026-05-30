package variant

import (
	"errors"
	"reflect"
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
			v, err := ToVariant(tc.input)
			if err != nil {
				t.Fatalf("ToVariant failed: %v", err)
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

func TestToVariantTypedNilPointers(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{name: "int32 pointer", input: (*int32)(nil)},
		{name: "array pointer", input: (*Array)(nil)},
		{name: "slice pointer", input: (*[]any)(nil)},
		{name: "map pointer", input: (*map[string]any)(nil)},
		{name: "error pointer", input: (*Error)(nil)},
		{name: "callchain pointer", input: (*CallChain)(nil)},
		{name: "reflect value pointer", input: (*reflect.Value)(nil)},
		{name: "variant pointer", input: (*Variant)(nil)},
		{name: "readable value pointer", input: (*convertTestValue)(nil)},
		{name: "reflect value holding nil pointer", input: reflect.ValueOf((*int32)(nil))},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v, err := ToVariant(tc.input)
			if err != nil {
				t.Fatalf("ToVariant failed: %v", err)
			}
			if v.TypeId != TypeId_Null {
				t.Fatalf("type id mismatch: got %d want %d", v.TypeId, TypeId_Null)
			}
		})
	}
}

func TestToVariantNilContainersKeepContainerType(t *testing.T) {
	var bs []byte
	v, err := ToVariant(bs)
	if err != nil {
		t.Fatalf("ToVariant nil []byte failed: %v", err)
	}
	if v.TypeId != TypeId_Bytes {
		t.Fatalf("nil []byte type id = %d, want %d", v.TypeId, TypeId_Bytes)
	}

	var m map[string]any
	v, err = ToVariant(m)
	if err != nil {
		t.Fatalf("ToVariant nil map failed: %v", err)
	}
	if v.TypeId != TypeId_Map {
		t.Fatalf("nil map type id = %d, want %d", v.TypeId, TypeId_Map)
	}
}

func TestToVariantTypedContainers(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		target reflect.Type
		want   any
	}{
		{
			name:   "typed slice",
			input:  []int{1, 2, 3},
			target: reflect.TypeFor[[]int](),
			want:   []int{1, 2, 3},
		},
		{
			name:   "typed array",
			input:  [2]string{"a", "b"},
			target: reflect.TypeFor[[2]string](),
			want:   [2]string{"a", "b"},
		},
		{
			name:   "typed map",
			input:  map[int]string{1: "one", 2: "two"},
			target: reflect.TypeFor[map[int]string](),
			want:   map[int]string{1: "one", 2: "two"},
		},
		{
			name:   "nested container",
			input:  map[string][]int{"nums": {1, 2, 3}},
			target: reflect.TypeFor[map[string][]int](),
			want:   map[string][]int{"nums": {1, 2, 3}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v, err := ToVariant(tc.input)
			if err != nil {
				t.Fatalf("ToVariant failed: %v", err)
			}

			got, err := v.ToNative(tc.target)
			if err != nil {
				t.Fatalf("ToNative failed: %v", err)
			}
			if !reflect.DeepEqual(got.Interface(), tc.want) {
				t.Fatalf("round trip = %#v, want %#v", got.Interface(), tc.want)
			}
		})
	}
}

func TestToVariantTypedContainerPointers(t *testing.T) {
	s := []int{1, 2, 3}
	v, err := ToVariant(&s)
	if err != nil {
		t.Fatalf("ToVariant *[]int failed: %v", err)
	}
	got, err := v.ToNative(reflect.TypeFor[[]int]())
	if err != nil {
		t.Fatalf("ToNative []int failed: %v", err)
	}
	if !reflect.DeepEqual(got.Interface(), s) {
		t.Fatalf("round trip slice = %#v, want %#v", got.Interface(), s)
	}

	m := map[int]string{1: "one"}
	v, err = ToVariant(&m)
	if err != nil {
		t.Fatalf("ToVariant *map[int]string failed: %v", err)
	}
	got, err = v.ToNative(reflect.TypeFor[map[int]string]())
	if err != nil {
		t.Fatalf("ToNative map[int]string failed: %v", err)
	}
	if !reflect.DeepEqual(got.Interface(), m) {
		t.Fatalf("round trip map = %#v, want %#v", got.Interface(), m)
	}
}

func TestToVariantInterfaceWrappedContainers(t *testing.T) {
	var slice any = []int{1, 2, 3}
	v, err := ToVariant(&slice)
	if err != nil {
		t.Fatalf("ToVariant *any slice failed: %v", err)
	}
	got, err := v.ToNative(reflect.TypeFor[[]int]())
	if err != nil {
		t.Fatalf("ToNative []int failed: %v", err)
	}
	if !reflect.DeepEqual(got.Interface(), []int{1, 2, 3}) {
		t.Fatalf("round trip interface slice = %#v", got.Interface())
	}

	var m any = map[int]string{1: "one"}
	v, err = ToVariant(&m)
	if err != nil {
		t.Fatalf("ToVariant *any map failed: %v", err)
	}
	got, err = v.ToNative(reflect.TypeFor[map[int]string]())
	if err != nil {
		t.Fatalf("ToNative map[int]string failed: %v", err)
	}
	if !reflect.DeepEqual(got.Interface(), map[int]string{1: "one"}) {
		t.Fatalf("round trip interface map = %#v", got.Interface())
	}
}

func TestToVariantContainersWithInterfaceValues(t *testing.T) {
	v, err := ToVariant([]any{[]int{1, 2}, []int{3}})
	if err != nil {
		t.Fatalf("ToVariant []any failed: %v", err)
	}
	got, err := v.ToNative(reflect.TypeFor[[][]int]())
	if err != nil {
		t.Fatalf("ToNative [][]int failed: %v", err)
	}
	if !reflect.DeepEqual(got.Interface(), [][]int{{1, 2}, {3}}) {
		t.Fatalf("converted [][]int = %#v", got.Interface())
	}

	v, err = ToVariant(map[string]any{"nums": []int{1, 2}, "nil": nil})
	if err != nil {
		t.Fatalf("ToVariant map[string]any failed: %v", err)
	}
	got, err = v.ToNative(reflect.TypeFor[map[string][]int]())
	if err != nil {
		t.Fatalf("ToNative map[string][]int failed: %v", err)
	}
	m := got.Interface().(map[string][]int)
	if !reflect.DeepEqual(m["nums"], []int{1, 2}) || m["nil"] != nil {
		t.Fatalf("converted map[string][]int = %#v", m)
	}
}

func TestToVariantCustomValuesInContainers(t *testing.T) {
	v1 := &convertTestValue{N: 1}
	v2 := &convertTestValue{N: 2}

	tests := []struct {
		name   string
		input  any
		target reflect.Type
		want   any
	}{
		{
			name:   "slice pointers",
			input:  []*convertTestValue{v1, v2},
			target: reflect.TypeFor[[]*convertTestValue](),
			want:   []*convertTestValue{v1, v2},
		},
		{
			name:   "array pointers",
			input:  [2]*convertTestValue{v1, v2},
			target: reflect.TypeFor[[2]*convertTestValue](),
			want:   [2]*convertTestValue{v1, v2},
		},
		{
			name:   "map pointers",
			input:  map[string]*convertTestValue{"one": v1, "two": v2},
			target: reflect.TypeFor[map[string]*convertTestValue](),
			want:   map[string]*convertTestValue{"one": v1, "two": v2},
		},
		{
			name:   "slice values",
			input:  []convertTestValue{{N: 1}, {N: 2}},
			target: reflect.TypeFor[[]convertTestValue](),
			want:   []convertTestValue{{N: 1}, {N: 2}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v, err := ToVariant(tc.input)
			if err != nil {
				t.Fatalf("ToVariant failed: %v", err)
			}
			if v.TypeId != TypeId_Array && v.TypeId != TypeId_Map {
				t.Fatalf("type id = %v, want Array or Map", v.TypeId)
			}

			got, err := v.ToNative(tc.target)
			if err != nil {
				t.Fatalf("ToNative failed: %v", err)
			}
			if !reflect.DeepEqual(got.Interface(), tc.want) {
				t.Fatalf("round trip = %#v, want %#v", got.Interface(), tc.want)
			}
		})
	}
}

func TestToVariantCustomValuesInUnaddressableContainersFail(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{name: "array values", input: [2]convertTestValue{{N: 1}, {N: 2}}},
		{name: "map values", input: map[string]convertTestValue{"one": {N: 1}, "two": {N: 2}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ToVariant(tc.input); !errors.Is(err, ErrInvalidCast) {
				t.Fatalf("ToVariant error = %v, want ErrInvalidCast", err)
			}
		})
	}
}

func TestToVariantContainerInvalidItems(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{name: "slice item", input: []complex64{1 + 2i}},
		{name: "array item", input: [1]complex64{1 + 2i}},
		{name: "map key", input: map[complex64]string{1 + 2i: "bad"}},
		{name: "map value", input: map[string]complex64{"bad": 1 + 2i}},
		{name: "nested slice item", input: [][]complex64{{1 + 2i}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ToVariant(tc.input); !errors.Is(err, ErrInvalidCast) {
				t.Fatalf("ToVariant error = %v, want ErrInvalidCast", err)
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
