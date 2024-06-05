package variant

import (
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/uid"
	"reflect"
)

// CustomCastVariant 自定义转换可变类型
var CustomCastVariant func(a any) (Variant, error)

// CastVariant 转换可变类型
func CastVariant(a any) (Variant, error) {
retry:
	switch v := a.(type) {
	case int:
		return MakeVariant(Int(v))
	case *int:
		return MakeVariant((*Int)(v))
	case int8:
		return MakeVariant(Int8(v))
	case *int8:
		return MakeVariant((*Int8)(v))
	case int16:
		return MakeVariant(Int16(v))
	case *int16:
		return MakeVariant((*Int16)(v))
	case int32:
		return MakeVariant(Int32(v))
	case *int32:
		return MakeVariant((*Int32)(v))
	case int64:
		return MakeVariant(Int64(v))
	case *int64:
		return MakeVariant((*Int64)(v))
	case uint:
		return MakeVariant(Uint(v))
	case *uint:
		return MakeVariant((*Uint)(v))
	case uint8:
		return MakeVariant(Uint8(v))
	case *uint8:
		return MakeVariant((*Uint8)(v))
	case uint16:
		return MakeVariant(Uint16(v))
	case *uint16:
		return MakeVariant((*Uint16)(v))
	case uint32:
		return MakeVariant(Uint32(v))
	case *uint32:
		return MakeVariant((*Uint32)(v))
	case uint64:
		return MakeVariant(Uint64(v))
	case *uint64:
		return MakeVariant((*Uint64)(v))
	case float32:
		return MakeVariant(Float(v))
	case *float32:
		return MakeVariant((*Float)(v))
	case float64:
		return MakeVariant(Double(v))
	case *float64:
		return MakeVariant((*Double)(v))
	case bool:
		return MakeVariant(Bool(v))
	case *bool:
		return MakeVariant((*Bool)(v))
	case []byte:
		return MakeVariant(Bytes(v))
	case *[]byte:
		return MakeVariant((*Bytes)(v))
	case string:
		return MakeVariant(String(v))
	case *string:
		return MakeVariant((*String)(v))
	case uid.Id:
		return MakeVariant(String(v))
	case *uid.Id:
		return MakeVariant((*String)(v))
	case nil:
		return MakeVariant(Null{})
	case Array:
		return MakeVariant(v)
	case *Array:
		return MakeVariant(v)
	case []any:
		arr, err := MakeArray(v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(arr)
	case *[]any:
		arr, err := MakeArray(*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(arr)
	case []reflect.Value:
		arr, err := MakeArray(v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(arr)
	case *[]reflect.Value:
		arr, err := MakeArray(*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(arr)
	case Map:
		return MakeVariant(v)
	case *Map:
		return MakeVariant(v)
	case map[string]any:
		m, err := MakeMapFromGoMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(m)
	case *map[string]any:
		m, err := MakeMapFromGoMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(m)
	case generic.SliceMap[string, any]:
		m, err := MakeMapFromSliceMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(m)
	case *generic.SliceMap[string, any]:
		m, err := MakeMapFromSliceMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(m)
	case Error:
		return MakeVariant(&v)
	case *Error:
		return MakeVariant(v)
	case error:
		return MakeVariant(MakeError(v))
	case CallChain:
		return MakeVariant(v)
	case *CallChain:
		return MakeVariant(*v)
	case reflect.Value:
		if !v.CanInterface() {
			return Variant{}, ErrNotVariant
		}
		a = v.Interface()
		goto retry
	case *reflect.Value:
		if !v.CanInterface() {
			return Variant{}, ErrNotVariant
		}
		a = v.Interface()
		goto retry
	case Variant:
		return v, nil
	case *Variant:
		return *v, nil
	case ValueReader:
		return MakeVariant(v)
	default:
		if CustomCastVariant != nil {
			return CustomCastVariant(a)
		}
		return Variant{}, ErrNotVariant
	}
}
