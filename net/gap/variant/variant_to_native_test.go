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
	v, err := ToVariant(int32(9))
	if err != nil {
		t.Fatalf("ToVariant failed: %v", err)
	}

	got, err := v.ToNative(reflect.TypeFor[int32]())
	if err != nil {
		t.Fatalf("ToNative int32 failed: %v", err)
	}
	if got.Interface().(int32) != 9 {
		t.Fatalf("converted int32 = %v, want 9", got.Interface())
	}

	if _, err := v.ToNative(reflect.TypeFor[complex64]()); !errors.Is(err, ErrInvalidCast) {
		t.Fatalf("ToNative complex64 error = %v, want ErrInvalidCast", err)
	}
}

func TestVariantConvertPointerRules(t *testing.T) {
	v, err := ToVariant(int32(9))
	if err != nil {
		t.Fatalf("ToVariant failed: %v", err)
	}

	if _, err := v.ToNative(reflect.TypeFor[*Int32]()); !errors.Is(err, ErrInvalidCast) {
		t.Fatalf("constructed T -> *T error = %v, want ErrInvalidCast", err)
	}

	decoded := decodeVariant(t, encodeVariant(t, v))

	gotPtr, err := decoded.ToNative(reflect.TypeFor[*Int32]())
	if err != nil {
		t.Fatalf("decoded *T -> *T failed: %v", err)
	}
	if *gotPtr.Interface().(*Int32) != Int32(9) {
		t.Fatalf("converted *Int32 = %v, want 9", gotPtr.Interface())
	}

	gotValue, err := decoded.ToNative(reflect.TypeFor[Int32]())
	if err != nil {
		t.Fatalf("decoded *T -> T failed: %v", err)
	}
	if gotValue.Interface().(Int32) != Int32(9) {
		t.Fatalf("converted Int32 = %v, want 9", gotValue.Interface())
	}
}

func TestVariantConvertDecodedArray(t *testing.T) {
	v, err := ToVariant([]any{"x", int32(3), true})
	if err != nil {
		t.Fatalf("ToVariant failed: %v", err)
	}
	decoded := decodeVariant(t, encodeVariant(t, v))

	got, err := decoded.ToNative(reflect.TypeFor[[]any]())
	if err != nil {
		t.Fatalf("ToNative []any failed: %v", err)
	}

	items := got.Interface().([]any)
	if len(items) != 3 {
		t.Fatalf("converted length = %d, want 3", len(items))
	}
	if items[0].(string) != "x" || items[1].(int32) != 3 || items[2].(bool) != true {
		t.Fatalf("converted items = %#v", items)
	}

	gotPtr, err := decoded.ToNative(reflect.TypeFor[*[]any]())
	if err != nil {
		t.Fatalf("ToNative *[]any failed: %v", err)
	}
	ptrItems := gotPtr.Interface().(*[]any)
	if len(*ptrItems) != 3 || (*ptrItems)[0].(string) != "x" || (*ptrItems)[1].(int32) != 3 || (*ptrItems)[2].(bool) != true {
		t.Fatalf("converted *[]any = %#v", ptrItems)
	}

	gotReflectValues, err := decoded.ToNative(reflect.TypeFor[[]reflect.Value]())
	if err != nil {
		t.Fatalf("ToNative []reflect.Value failed: %v", err)
	}
	reflectValues := gotReflectValues.Interface().([]reflect.Value)
	if len(reflectValues) != 3 || reflectValues[0].Interface().(string) != "x" || reflectValues[1].Interface().(int32) != 3 || reflectValues[2].Interface().(bool) != true {
		t.Fatalf("converted []reflect.Value = %#v", reflectValues)
	}

	gotReflectValuesPtr, err := decoded.ToNative(reflect.TypeFor[*[]reflect.Value]())
	if err != nil {
		t.Fatalf("ToNative *[]reflect.Value failed: %v", err)
	}
	if gotReflectValuesPtr.Type() != reflect.TypeFor[*[]reflect.Value]() {
		t.Fatalf("converted type = %v, want %v", gotReflectValuesPtr.Type(), reflect.TypeFor[*[]reflect.Value]())
	}
}

func TestVariantConvertDecodedMap(t *testing.T) {
	v, err := ToVariant(map[string]any{"name": "demo", "count": int32(2)})
	if err != nil {
		t.Fatalf("ToVariant failed: %v", err)
	}
	decoded := decodeVariant(t, encodeVariant(t, v))

	gotMap, err := decoded.ToNative(reflect.TypeFor[map[string]any]())
	if err != nil {
		t.Fatalf("ToNative map[string]any failed: %v", err)
	}
	m := gotMap.Interface().(map[string]any)
	if m["name"].(string) != "demo" || m["count"].(int32) != 2 {
		t.Fatalf("converted map = %#v", m)
	}

	gotMapPtr, err := decoded.ToNative(reflect.TypeFor[*map[string]any]())
	if err != nil {
		t.Fatalf("ToNative *map[string]any failed: %v", err)
	}
	mapPtr := gotMapPtr.Interface().(*map[string]any)
	if (*mapPtr)["name"].(string) != "demo" || (*mapPtr)["count"].(int32) != 2 {
		t.Fatalf("converted *map[string]any = %#v", mapPtr)
	}

	gotSliceMap, err := decoded.ToNative(reflect.TypeFor[generic.SliceMap[string, any]]())
	if err != nil {
		t.Fatalf("ToNative SliceMap failed: %v", err)
	}
	if gotSliceMap.Interface().(generic.SliceMap[string, any]).Len() != 2 {
		t.Fatalf("converted SliceMap length = %d, want 2", gotSliceMap.Len())
	}

	gotSliceMapPtr, err := decoded.ToNative(reflect.TypeFor[*generic.SliceMap[string, any]]())
	if err != nil {
		t.Fatalf("ToNative *SliceMap failed: %v", err)
	}
	if gotSliceMapPtr.Type() != reflect.TypeFor[*generic.SliceMap[string, any]]() {
		t.Fatalf("converted type = %v, want %v", gotSliceMapPtr.Type(), reflect.TypeFor[*generic.SliceMap[string, any]]())
	}

	gotUnorderedSliceMap, err := decoded.ToNative(reflect.TypeFor[generic.UnorderedSliceMap[string, any]]())
	if err != nil {
		t.Fatalf("ToNative UnorderedSliceMap failed: %v", err)
	}
	unorderedSliceMap := gotUnorderedSliceMap.Interface().(generic.UnorderedSliceMap[string, any])
	if unorderedSliceMap.Len() != 2 {
		t.Fatalf("converted UnorderedSliceMap length = %d, want 2", unorderedSliceMap.Len())
	}
	if name, ok := unorderedSliceMap.Get("name"); !ok || name.(string) != "demo" {
		t.Fatalf("converted UnorderedSliceMap name = %#v, exists = %t", name, ok)
	}
	if count, ok := unorderedSliceMap.Get("count"); !ok || count.(int32) != 2 {
		t.Fatalf("converted UnorderedSliceMap count = %#v, exists = %t", count, ok)
	}

	gotUnorderedSliceMapPtr, err := decoded.ToNative(reflect.TypeFor[*generic.UnorderedSliceMap[string, any]]())
	if err != nil {
		t.Fatalf("ToNative *UnorderedSliceMap failed: %v", err)
	}
	if gotUnorderedSliceMapPtr.Type() != reflect.TypeFor[*generic.UnorderedSliceMap[string, any]]() {
		t.Fatalf("converted type = %v, want %v", gotUnorderedSliceMapPtr.Type(), reflect.TypeFor[*generic.UnorderedSliceMap[string, any]]())
	}
}

func TestVariantConvertMapWithNonStringKeyToStringAnyFails(t *testing.T) {
	v, err := ToVariant(map[int]string{1: "one"})
	if err != nil {
		t.Fatalf("ToVariant failed: %v", err)
	}

	if _, err := v.ToNative(reflect.TypeFor[map[string]any]()); !errors.Is(err, ErrInvalidCast) {
		t.Fatalf("ToNative map[string]any error = %v, want ErrInvalidCast", err)
	}
	if _, err := v.ToNative(reflect.TypeFor[generic.SliceMap[string, any]]()); !errors.Is(err, ErrInvalidCast) {
		t.Fatalf("ToNative SliceMap[string, any] error = %v, want ErrInvalidCast", err)
	}
	if _, err := v.ToNative(reflect.TypeFor[generic.UnorderedSliceMap[string, any]]()); !errors.Is(err, ErrInvalidCast) {
		t.Fatalf("ToNative UnorderedSliceMap[string, any] error = %v, want ErrInvalidCast", err)
	}
}

func TestVariantConvertNullToNilContainers(t *testing.T) {
	v, err := ToVariant(nil)
	if err != nil {
		t.Fatalf("ToVariant failed: %v", err)
	}

	got, err := v.ToNative(reflect.TypeFor[map[string]any]())
	if err != nil {
		t.Fatalf("ToNative null map failed: %v", err)
	}
	if !got.IsNil() {
		t.Fatalf("converted null map is not nil: %#v", got.Interface())
	}
}

func TestVariantConvertReflectValueFallback(t *testing.T) {
	v, err := ToVariant(int32(9))
	if err != nil {
		t.Fatalf("ToVariant failed: %v", err)
	}

	got, err := v.ToNative(reflect.TypeFor[reflect.Value]())
	if err != nil {
		t.Fatalf("ToNative reflect.Value failed: %v", err)
	}
	if got.Type() != reflect.TypeFor[Int32]() {
		t.Fatalf("reflect.Value type = %v, want %v", got.Type(), reflect.TypeFor[Int32]())
	}

	gotPtr, err := v.ToNative(reflect.TypeFor[*reflect.Value]())
	if err != nil {
		t.Fatalf("ToNative *reflect.Value failed: %v", err)
	}
	if gotPtr.Type() != reflect.TypeFor[*reflect.Value]() {
		t.Fatalf("converted type = %v, want %v", gotPtr.Type(), reflect.TypeFor[*reflect.Value]())
	}
	if gotPtr.Interface().(*reflect.Value).Type() != reflect.TypeFor[Int32]() {
		t.Fatalf("*reflect.Value type = %v, want %v", gotPtr.Interface().(*reflect.Value).Type(), reflect.TypeFor[Int32]())
	}
}

func TestVariantConvertValueInterfaceTargets(t *testing.T) {
	v, err := ToVariant(int32(9))
	if err != nil {
		t.Fatalf("ToVariant failed: %v", err)
	}

	gotValue, err := v.ToNative(reflect.TypeFor[Value]())
	if err != nil {
		t.Fatalf("ToNative Value failed: %v", err)
	}
	if gotValue.Interface().(Value).TypeID() != TypeID_Int32 {
		t.Fatalf("converted Value TypeID = %v, want %v", gotValue.Interface().(Value).TypeID(), TypeID_Int32)
	}

	gotReadableValue, err := v.ToNative(reflect.TypeFor[ReadableValue]())
	if err != nil {
		t.Fatalf("ToNative ReadableValue failed: %v", err)
	}
	if gotReadableValue.Interface().(ReadableValue).TypeID() != TypeID_Int32 {
		t.Fatalf("converted ReadableValue TypeID = %v, want %v", gotReadableValue.Interface().(ReadableValue).TypeID(), TypeID_Int32)
	}

	arr, err := ToVariant([2]int32{1, 2})
	if err != nil {
		t.Fatalf("ToVariant array failed: %v", err)
	}
	gotSlice, err := arr.ToNative(reflect.TypeFor[[]Value]())
	if err != nil {
		t.Fatalf("ToNative []Value failed: %v", err)
	}
	values := gotSlice.Interface().([]Value)
	if len(values) != 2 || values[0].TypeID() != TypeID_Int32 || values[1].TypeID() != TypeID_Int32 {
		t.Fatalf("converted []Value = %#v", values)
	}

	nullVariant, err := ToVariant(nil)
	if err != nil {
		t.Fatalf("ToVariant nil failed: %v", err)
	}
	gotNull, err := nullVariant.ToNative(reflect.TypeFor[Value]())
	if err != nil {
		t.Fatalf("ToNative null Value failed: %v", err)
	}
	if gotNull.Interface().(Value).TypeID() != TypeID_Null {
		t.Fatalf("converted null Value TypeID = %v, want %v", gotNull.Interface().(Value).TypeID(), TypeID_Null)
	}
}

func TestVariantConvertVariantPointer(t *testing.T) {
	v, err := ToVariant(int32(9))
	if err != nil {
		t.Fatalf("ToVariant failed: %v", err)
	}

	gotValue, err := v.ToNative(reflect.TypeFor[Variant]())
	if err != nil {
		t.Fatalf("ToNative Variant failed: %v", err)
	}
	if gotValue.Type() != reflect.TypeFor[Variant]() {
		t.Fatalf("converted type = %v, want %v", gotValue.Type(), reflect.TypeFor[Variant]())
	}

	got, err := v.ToNative(reflect.TypeFor[*Variant]())
	if err != nil {
		t.Fatalf("ToNative *Variant failed: %v", err)
	}
	if got.Type() != reflect.TypeFor[*Variant]() {
		t.Fatalf("converted type = %v, want %v", got.Type(), reflect.TypeFor[*Variant]())
	}
	if got.Interface().(*Variant).TypeID != TypeID_Int32 {
		t.Fatalf("converted Variant TypeID = %v, want %v", got.Interface().(*Variant).TypeID, TypeID_Int32)
	}
}

func TestVariantConvertTypedSlice(t *testing.T) {
	v, err := ToVariant([]any{1, 2, 3})
	if err != nil {
		t.Fatalf("ToVariant failed: %v", err)
	}
	decoded := decodeVariant(t, encodeVariant(t, v))

	got, err := decoded.ToNative(reflect.TypeFor[[]int]())
	if err != nil {
		t.Fatalf("ToNative []int failed: %v", err)
	}

	items := got.Interface().([]int)
	if !reflect.DeepEqual(items, []int{1, 2, 3}) {
		t.Fatalf("converted []int = %#v", items)
	}
}

func TestVariantConvertNestedTypedSlice(t *testing.T) {
	v, err := ToVariant([]any{
		[]any{int32(1), int32(2)},
		[]any{int32(3), int32(4)},
	})
	if err != nil {
		t.Fatalf("ToVariant failed: %v", err)
	}
	decoded := decodeVariant(t, encodeVariant(t, v))

	got, err := decoded.ToNative(reflect.TypeFor[[][]int32]())
	if err != nil {
		t.Fatalf("ToNative [][]int32 failed: %v", err)
	}

	items := got.Interface().([][]int32)
	if !reflect.DeepEqual(items, [][]int32{{1, 2}, {3, 4}}) {
		t.Fatalf("converted [][]int32 = %#v", items)
	}
}

func TestVariantConvertTypedArray(t *testing.T) {
	v, err := ToVariant([]any{"a", "b"})
	if err != nil {
		t.Fatalf("ToVariant failed: %v", err)
	}
	decoded := decodeVariant(t, encodeVariant(t, v))

	got, err := decoded.ToNative(reflect.TypeFor[[2]string]())
	if err != nil {
		t.Fatalf("ToNative [2]string failed: %v", err)
	}

	items := got.Interface().([2]string)
	if items != [2]string{"a", "b"} {
		t.Fatalf("converted [2]string = %#v", items)
	}

	if _, err := decoded.ToNative(reflect.TypeFor[[3]string]()); !errors.Is(err, ErrInvalidCast) {
		t.Fatalf("ToNative [3]string error = %v, want ErrInvalidCast", err)
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

	got, err := decoded.ToNative(reflect.TypeFor[map[string]int]())
	if err != nil {
		t.Fatalf("ToNative map[string]int failed: %v", err)
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

	got, err := decoded.ToNative(reflect.TypeFor[map[string]map[string]int]())
	if err != nil {
		t.Fatalf("ToNative nested map failed: %v", err)
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

func (convertTestValue) TypeID() TypeID {
	return TypeID_Customize + 10001
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

	gotPtrs, err := v.ToNative(reflect.TypeFor[[]*convertTestValue]())
	if err != nil {
		t.Fatalf("ToNative []*convertTestValue failed: %v", err)
	}
	ptrs := gotPtrs.Interface().([]*convertTestValue)
	if len(ptrs) != 2 || ptrs[0] != v1 || ptrs[1] != v2 {
		t.Fatalf("converted []*convertTestValue = %#v", ptrs)
	}

	gotValues, err := v.ToNative(reflect.TypeFor[[]convertTestValue]())
	if err != nil {
		t.Fatalf("ToNative []convertTestValue failed: %v", err)
	}
	values := gotValues.Interface().([]convertTestValue)
	if !reflect.DeepEqual(values, []convertTestValue{{N: 1}, {N: 2}}) {
		t.Fatalf("converted []convertTestValue = %#v", values)
	}
}

func TestVariantConvertCustomValueInTypedArray(t *testing.T) {
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

	gotPtrs, err := v.ToNative(reflect.TypeFor[[2]*convertTestValue]())
	if err != nil {
		t.Fatalf("ToNative [2]*convertTestValue failed: %v", err)
	}
	ptrs := gotPtrs.Interface().([2]*convertTestValue)
	if ptrs[0] != v1 || ptrs[1] != v2 {
		t.Fatalf("converted [2]*convertTestValue = %#v", ptrs)
	}

	gotValues, err := v.ToNative(reflect.TypeFor[[2]convertTestValue]())
	if err != nil {
		t.Fatalf("ToNative [2]convertTestValue failed: %v", err)
	}
	values := gotValues.Interface().([2]convertTestValue)
	if values != [2]convertTestValue{{N: 1}, {N: 2}} {
		t.Fatalf("converted [2]convertTestValue = %#v", values)
	}
}

func TestVariantConvertCustomValueInTypedMap(t *testing.T) {
	v1 := &convertTestValue{N: 1}
	v2 := &convertTestValue{N: 2}

	m, err := NewMapFromGoMap(map[string]*convertTestValue{"one": v1, "two": v2})
	if err != nil {
		t.Fatalf("NewMapFromGoMap failed: %v", err)
	}
	v, err := NewVariant(m)
	if err != nil {
		t.Fatalf("NewVariant failed: %v", err)
	}

	gotPtrs, err := v.ToNative(reflect.TypeFor[map[string]*convertTestValue]())
	if err != nil {
		t.Fatalf("ToNative map[string]*convertTestValue failed: %v", err)
	}
	ptrs := gotPtrs.Interface().(map[string]*convertTestValue)
	if ptrs["one"] != v1 || ptrs["two"] != v2 {
		t.Fatalf("converted map[string]*convertTestValue = %#v", ptrs)
	}

	gotValues, err := v.ToNative(reflect.TypeFor[map[string]convertTestValue]())
	if err != nil {
		t.Fatalf("ToNative map[string]convertTestValue failed: %v", err)
	}
	values := gotValues.Interface().(map[string]convertTestValue)
	if !reflect.DeepEqual(values, map[string]convertTestValue{"one": {N: 1}, "two": {N: 2}}) {
		t.Fatalf("converted map[string]convertTestValue = %#v", values)
	}
}

func TestVariantConvertCustomValueTargets(t *testing.T) {
	src := &convertTestValue{N: 7}
	v, err := ToVariant(src)
	if err != nil {
		t.Fatalf("ToVariant custom value failed: %v", err)
	}

	gotPtr, err := v.ToNative(reflect.TypeFor[*convertTestValue]())
	if err != nil {
		t.Fatalf("ToNative *convertTestValue failed: %v", err)
	}
	if gotPtr.Interface().(*convertTestValue) != src {
		t.Fatalf("converted pointer = %p, want %p", gotPtr.Interface().(*convertTestValue), src)
	}

	gotValue, err := v.ToNative(reflect.TypeFor[convertTestValue]())
	if err != nil {
		t.Fatalf("ToNative convertTestValue failed: %v", err)
	}
	if gotValue.Interface().(convertTestValue) != (convertTestValue{N: 7}) {
		t.Fatalf("converted value = %#v, want N=7", gotValue.Interface())
	}

	gotReadable, err := v.ToNative(reflect.TypeFor[ReadableValue]())
	if err != nil {
		t.Fatalf("ToNative ReadableValue failed: %v", err)
	}
	if gotReadable.Interface().(ReadableValue).TypeID() != src.TypeID() {
		t.Fatalf("converted ReadableValue TypeID = %v, want %v", gotReadable.Interface().(ReadableValue).TypeID(), src.TypeID())
	}

	gotValueInterface, err := v.ToNative(reflect.TypeFor[Value]())
	if err != nil {
		t.Fatalf("ToNative Value failed: %v", err)
	}
	if gotValueInterface.Interface().(Value).TypeID() != src.TypeID() {
		t.Fatalf("converted Value TypeID = %v, want %v", gotValueInterface.Interface().(Value).TypeID(), src.TypeID())
	}
}
