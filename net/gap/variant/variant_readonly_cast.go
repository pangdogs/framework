package variant

import (
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/uid"
	"reflect"
)

// CastVariantReadonly 转换只读可变类型
func CastVariantReadonly(a any) (Variant, error) {
retry:
	switch v := a.(type) {
	case int:
		return MakeVariantReadonly(Int(v))
	case *int:
		return MakeVariantReadonly((*Int)(v))
	case int8:
		return MakeVariantReadonly(Int8(v))
	case *int8:
		return MakeVariantReadonly((*Int8)(v))
	case int16:
		return MakeVariantReadonly(Int16(v))
	case *int16:
		return MakeVariantReadonly((*Int16)(v))
	case int32:
		return MakeVariantReadonly(Int32(v))
	case *int32:
		return MakeVariantReadonly((*Int32)(v))
	case int64:
		return MakeVariantReadonly(Int64(v))
	case *int64:
		return MakeVariantReadonly((*Int64)(v))
	case uint:
		return MakeVariantReadonly(Uint(v))
	case *uint:
		return MakeVariantReadonly((*Uint)(v))
	case uint8:
		return MakeVariantReadonly(Uint8(v))
	case *uint8:
		return MakeVariantReadonly((*Uint8)(v))
	case uint16:
		return MakeVariantReadonly(Uint16(v))
	case *uint16:
		return MakeVariantReadonly((*Uint16)(v))
	case uint32:
		return MakeVariantReadonly(Uint32(v))
	case *uint32:
		return MakeVariantReadonly((*Uint32)(v))
	case uint64:
		return MakeVariantReadonly(Uint64(v))
	case *uint64:
		return MakeVariantReadonly((*Uint64)(v))
	case float32:
		return MakeVariantReadonly(Float(v))
	case *float32:
		return MakeVariantReadonly((*Float)(v))
	case float64:
		return MakeVariantReadonly(Double(v))
	case *float64:
		return MakeVariantReadonly((*Double)(v))
	case bool:
		return MakeVariantReadonly(Bool(v))
	case *bool:
		return MakeVariantReadonly((*Bool)(v))
	case []byte:
		return MakeVariantReadonly(Bytes(v))
	case *[]byte:
		return MakeVariantReadonly((*Bytes)(v))
	case string:
		return MakeVariantReadonly(String(v))
	case *string:
		return MakeVariantReadonly((*String)(v))
	case uid.Id:
		return MakeVariantReadonly(String(v))
	case *uid.Id:
		return MakeVariantReadonly((*String)(v))
	case nil:
		return MakeVariantReadonly(Null{})
	case Array:
		return MakeVariantReadonly(v)
	case *Array:
		return MakeVariantReadonly(*v)
	case []any:
		arr, err := MakeArrayReadonly(v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariantReadonly(arr)
	case *[]any:
		arr, err := MakeArrayReadonly(*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariantReadonly(arr)
	case []reflect.Value:
		arr, err := MakeArrayReadonly(v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariantReadonly(arr)
	case *[]reflect.Value:
		arr, err := MakeArrayReadonly(*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariantReadonly(arr)
	case Map:
		return MakeVariantReadonly(v)
	case *Map:
		return MakeVariantReadonly(*v)
	case map[string]any:
		m, err := MakeMapReadonlyFromGoMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariantReadonly(m)
	case *map[string]any:
		m, err := MakeMapReadonlyFromGoMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariantReadonly(m)
	case generic.SliceMap[string, any]:
		m, err := MakeMapReadonlyFromSliceMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariantReadonly(m)
	case *generic.SliceMap[string, any]:
		m, err := MakeMapReadonlyFromSliceMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariantReadonly(m)
	case generic.UnorderedSliceMap[string, any]:
		m, err := MakeMapReadonlyFromUnorderedSliceMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariantReadonly(m)
	case *generic.UnorderedSliceMap[string, any]:
		m, err := MakeMapReadonlyFromUnorderedSliceMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariantReadonly(m)
	case Error:
		return MakeVariantReadonly(&v)
	case *Error:
		return MakeVariantReadonly(v)
	case error:
		return MakeVariantReadonly(MakeError(v))
	case CallChain:
		return MakeVariantReadonly(v)
	case *CallChain:
		return MakeVariantReadonly(*v)
	case reflect.Value:
		if !v.CanInterface() {
			return Variant{}, ErrInvalidCast
		}
		a = v.Interface()
		goto retry
	case *reflect.Value:
		if !v.CanInterface() {
			return Variant{}, ErrInvalidCast
		}
		a = v.Interface()
		goto retry
	case Variant:
		return v, nil
	case *Variant:
		return *v, nil
	case ValueReader:
		return MakeVariantReadonly(v)
	default:
		return Variant{}, ErrInvalidCast
	}
}
