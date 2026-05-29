package variant

import (
	"errors"
	"io"
	"reflect"
	"testing"

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/utils/binaryutil"
)

func TestVariantConvertScalars(t *testing.T) {
	v, err := CastVariant(int32(9))
	if err != nil {
		t.Fatalf("CastVariant failed: %v", err)
	}

	got, err := v.Convert(reflect.TypeFor[int32]())
	if err != nil {
		t.Fatalf("Convert int32 failed: %v", err)
	}
	if got.Interface().(int32) != 9 {
		t.Fatalf("converted int32 = %v, want 9", got.Interface())
	}

	if _, err := v.Convert(reflect.TypeFor[complex64]()); !errors.Is(err, ErrInvalidCast) {
		t.Fatalf("Convert complex64 error = %v, want ErrInvalidCast", err)
	}
}

func TestVariantConvertPointerRules(t *testing.T) {
	v, err := CastVariant(int32(9))
	if err != nil {
		t.Fatalf("CastVariant failed: %v", err)
	}

	if _, err := v.Convert(reflect.TypeFor[*Int32]()); !errors.Is(err, ErrInvalidCast) {
		t.Fatalf("constructed T -> *T error = %v, want ErrInvalidCast", err)
	}

	decoded := decodeVariant(t, encodeVariant(t, v))

	gotPtr, err := decoded.Convert(reflect.TypeFor[*Int32]())
	if err != nil {
		t.Fatalf("decoded *T -> *T failed: %v", err)
	}
	if *gotPtr.Interface().(*Int32) != Int32(9) {
		t.Fatalf("converted *Int32 = %v, want 9", gotPtr.Interface())
	}

	gotValue, err := decoded.Convert(reflect.TypeFor[Int32]())
	if err != nil {
		t.Fatalf("decoded *T -> T failed: %v", err)
	}
	if gotValue.Interface().(Int32) != Int32(9) {
		t.Fatalf("converted Int32 = %v, want 9", gotValue.Interface())
	}
}

func TestVariantConvertDecodedArray(t *testing.T) {
	v, err := CastVariant([]any{"x", int32(3), true})
	if err != nil {
		t.Fatalf("CastVariant failed: %v", err)
	}
	decoded := decodeVariant(t, encodeVariant(t, v))

	got, err := decoded.Convert(reflect.TypeFor[[]any]())
	if err != nil {
		t.Fatalf("Convert []any failed: %v", err)
	}

	items := got.Interface().([]any)
	if len(items) != 3 {
		t.Fatalf("converted length = %d, want 3", len(items))
	}
	if items[0].(string) != "x" || items[1].(int32) != 3 || items[2].(bool) != true {
		t.Fatalf("converted items = %#v", items)
	}
}

func TestVariantConvertDecodedMap(t *testing.T) {
	v, err := CastVariant(map[string]any{"name": "demo", "count": int32(2)})
	if err != nil {
		t.Fatalf("CastVariant failed: %v", err)
	}
	decoded := decodeVariant(t, encodeVariant(t, v))

	gotMap, err := decoded.Convert(reflect.TypeFor[map[string]any]())
	if err != nil {
		t.Fatalf("Convert map[string]any failed: %v", err)
	}
	m := gotMap.Interface().(map[string]any)
	if m["name"].(string) != "demo" || m["count"].(int32) != 2 {
		t.Fatalf("converted map = %#v", m)
	}

	gotSliceMap, err := decoded.Convert(reflect.TypeFor[generic.SliceMap[string, any]]())
	if err != nil {
		t.Fatalf("Convert SliceMap failed: %v", err)
	}
	if gotSliceMap.Interface().(generic.SliceMap[string, any]).Len() != 2 {
		t.Fatalf("converted SliceMap length = %d, want 2", gotSliceMap.Len())
	}
}

func TestVariantConvertNullToNilContainers(t *testing.T) {
	v, err := CastVariant(nil)
	if err != nil {
		t.Fatalf("CastVariant failed: %v", err)
	}

	got, err := v.Convert(reflect.TypeFor[map[string]any]())
	if err != nil {
		t.Fatalf("Convert null map failed: %v", err)
	}
	if !got.IsNil() {
		t.Fatalf("converted null map is not nil: %#v", got.Interface())
	}
}

func TestVariantConvertReflectValueFallback(t *testing.T) {
	v, err := CastVariant(int32(9))
	if err != nil {
		t.Fatalf("CastVariant failed: %v", err)
	}

	got, err := v.Convert(reflect.TypeFor[reflect.Value]())
	if err != nil {
		t.Fatalf("Convert reflect.Value failed: %v", err)
	}
	if got.Type() != reflect.TypeFor[Int32]() {
		t.Fatalf("reflect.Value type = %v, want %v", got.Type(), reflect.TypeFor[Int32]())
	}
}

func TestVariantConvertTypedSlice(t *testing.T) {
	v, err := CastVariant([]any{1, 2, 3})
	if err != nil {
		t.Fatalf("CastVariant failed: %v", err)
	}
	decoded := decodeVariant(t, encodeVariant(t, v))

	got, err := decoded.Convert(reflect.TypeFor[[]int]())
	if err != nil {
		t.Fatalf("Convert []int failed: %v", err)
	}

	items := got.Interface().([]int)
	if !reflect.DeepEqual(items, []int{1, 2, 3}) {
		t.Fatalf("converted []int = %#v", items)
	}
}

func TestVariantConvertNestedTypedSlice(t *testing.T) {
	v, err := CastVariant([]any{
		[]any{int32(1), int32(2)},
		[]any{int32(3), int32(4)},
	})
	if err != nil {
		t.Fatalf("CastVariant failed: %v", err)
	}
	decoded := decodeVariant(t, encodeVariant(t, v))

	got, err := decoded.Convert(reflect.TypeFor[[][]int32]())
	if err != nil {
		t.Fatalf("Convert [][]int32 failed: %v", err)
	}

	items := got.Interface().([][]int32)
	if !reflect.DeepEqual(items, [][]int32{{1, 2}, {3, 4}}) {
		t.Fatalf("converted [][]int32 = %#v", items)
	}
}

func TestVariantConvertTypedArray(t *testing.T) {
	v, err := CastVariant([]any{"a", "b"})
	if err != nil {
		t.Fatalf("CastVariant failed: %v", err)
	}
	decoded := decodeVariant(t, encodeVariant(t, v))

	got, err := decoded.Convert(reflect.TypeFor[[2]string]())
	if err != nil {
		t.Fatalf("Convert [2]string failed: %v", err)
	}

	items := got.Interface().([2]string)
	if items != [2]string{"a", "b"} {
		t.Fatalf("converted [2]string = %#v", items)
	}

	if _, err := decoded.Convert(reflect.TypeFor[[3]string]()); !errors.Is(err, ErrInvalidCast) {
		t.Fatalf("Convert [3]string error = %v, want ErrInvalidCast", err)
	}
}

func TestVariantConvertTypedMap(t *testing.T) {
	m, err := NewMapFromGoMap(map[string]int{"one": 1, "two": 2})
	if err != nil {
		t.Fatalf("NewMapFromGoMap failed: %v", err)
	}
	v, err := NewVariant(m)
	if err != nil {
		t.Fatalf("NewVariant failed: %v", err)
	}
	decoded := decodeVariant(t, encodeVariant(t, v))

	got, err := decoded.Convert(reflect.TypeFor[map[string]int]())
	if err != nil {
		t.Fatalf("Convert map[string]int failed: %v", err)
	}

	items := got.Interface().(map[string]int)
	if !reflect.DeepEqual(items, map[string]int{"one": 1, "two": 2}) {
		t.Fatalf("converted map[string]int = %#v", items)
	}
}

func TestVariantConvertNestedTypedMap(t *testing.T) {
	inner, err := NewMapFromGoMap(map[string]int{"x": 1})
	if err != nil {
		t.Fatalf("NewMapFromGoMap inner failed: %v", err)
	}
	outer, err := NewMapFromGoMap(map[string]any{"inner": inner})
	if err != nil {
		t.Fatalf("NewMapFromGoMap outer failed: %v", err)
	}
	v, err := NewVariant(outer)
	if err != nil {
		t.Fatalf("NewVariant failed: %v", err)
	}
	decoded := decodeVariant(t, encodeVariant(t, v))

	got, err := decoded.Convert(reflect.TypeFor[map[string]map[string]int]())
	if err != nil {
		t.Fatalf("Convert nested map failed: %v", err)
	}

	items := got.Interface().(map[string]map[string]int)
	if !reflect.DeepEqual(items, map[string]map[string]int{"inner": map[string]int{"x": 1}}) {
		t.Fatalf("converted nested map = %#v", items)
	}
}

type convertTestValue struct {
	N int32
}

func (v convertTestValue) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteInt32(v.N); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

func (v *convertTestValue) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	n, err := bs.ReadInt32()
	if err != nil {
		return bs.BytesRead(), err
	}
	v.N = n
	return bs.BytesRead(), nil
}

func (convertTestValue) Size() int {
	return binaryutil.SizeofInt32
}

func (convertTestValue) TypeId() TypeId {
	return TypeId_Customize + 10001
}

func (v *convertTestValue) Indirect() any {
	return v
}

func TestVariantConvertCustomValueInTypedSlice(t *testing.T) {
	v1 := &convertTestValue{N: 1}
	v2 := &convertTestValue{N: 2}

	arr, err := NewArray([]any{v1, v2})
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}
	v, err := NewVariant(arr)
	if err != nil {
		t.Fatalf("NewVariant failed: %v", err)
	}

	gotPtrs, err := v.Convert(reflect.TypeFor[[]*convertTestValue]())
	if err != nil {
		t.Fatalf("Convert []*convertTestValue failed: %v", err)
	}
	ptrs := gotPtrs.Interface().([]*convertTestValue)
	if len(ptrs) != 2 || ptrs[0] != v1 || ptrs[1] != v2 {
		t.Fatalf("converted []*convertTestValue = %#v", ptrs)
	}

	gotValues, err := v.Convert(reflect.TypeFor[[]convertTestValue]())
	if err != nil {
		t.Fatalf("Convert []convertTestValue failed: %v", err)
	}
	values := gotValues.Interface().([]convertTestValue)
	if !reflect.DeepEqual(values, []convertTestValue{{N: 1}, {N: 2}}) {
		t.Fatalf("converted []convertTestValue = %#v", values)
	}
}
