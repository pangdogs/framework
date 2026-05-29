package variant

import (
	"errors"
	"reflect"
	"testing"

	"git.golaxy.org/core/utils/generic"
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
