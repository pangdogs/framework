package variant

import (
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/uid"
	"reflect"
)

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
		return MakeVariant(v)
	default:
		return Variant{}, ErrInvalidCast
	}
}

var (
	sliceAnyRT          = reflect.TypeFor[[]any]()
	sliceRVRT           = reflect.TypeFor[[]reflect.Value]()
	mapStringAnyRT      = reflect.TypeFor[map[string]any]()
	sliceMapStringAnyRT = reflect.TypeFor[generic.SliceMap[string, any]]()
	rvRT                = reflect.TypeFor[reflect.Value]()
	variantRT           = reflect.TypeFor[Variant]()
)

// CastReflected 转换反射类型
func CastReflected(v Variant, rt reflect.Type) (reflect.Value, error) {
	if !v.Reflected.IsValid() {
		return reflect.Value{}, ErrInvalidCast
	}

	switch rt {
	case sliceAnyRT:
		switch v.TypeId {
		case TypeId_Null:
			return reflect.Zero(rt), nil

		case TypeId_Array:
			arr := *v.Value.(*Array)

			rv := make([]any, 0, len(arr))
			for _, it := range arr {
				rv = append(rv, it.Value.Indirect())
			}
			return reflect.ValueOf(rv), nil

		default:
			return reflect.Value{}, ErrInvalidCast
		}
	case reflect.PointerTo(rt):
		switch v.TypeId {
		case TypeId_Null:
			return reflect.Zero(rt), nil

		case TypeId_Array:
			arr := *v.Value.(*Array)

			rv := make([]any, 0, len(arr))
			for _, it := range arr {
				rv = append(rv, it.Value.Indirect())
			}
			return reflect.ValueOf(&rv), nil

		default:
			return reflect.Value{}, ErrInvalidCast
		}
	case sliceRVRT:
		switch v.TypeId {
		case TypeId_Null:
			return reflect.Zero(rt), nil

		case TypeId_Array:
			arr := *v.Value.(*Array)

			rv := make([]reflect.Value, 0, len(arr))
			for _, it := range arr {
				rv = append(rv, it.Reflected.Elem())
			}
			return reflect.ValueOf(rv), nil

		default:
			return reflect.Value{}, ErrInvalidCast
		}
	case reflect.PointerTo(sliceRVRT):
		switch v.TypeId {
		case TypeId_Null:
			return reflect.Zero(rt), nil

		case TypeId_Array:
			arr := *v.Value.(*Array)

			rv := make([]reflect.Value, 0, len(arr))
			for _, it := range arr {
				rv = append(rv, it.Reflected.Elem())
			}
			return reflect.ValueOf(&rv), nil

		default:
			return reflect.Value{}, ErrInvalidCast
		}
	case mapStringAnyRT:
		switch v.TypeId {
		case TypeId_Null:
			return reflect.Zero(rt), nil

		case TypeId_Array:
			m := v.Value.(*Map).CastSliceMap()

			rv := make(map[string]any, len(m))
			for _, kv := range m {
				if kv.K.TypeId != TypeId_String {
					return reflect.Value{}, ErrInvalidCast
				}
				rv[kv.K.Value.Indirect().(string)] = kv.V.Value.Indirect()
			}
			return reflect.ValueOf(rv), nil

		default:
			return reflect.Value{}, ErrInvalidCast
		}
	case reflect.PointerTo(mapStringAnyRT):
		switch v.TypeId {
		case TypeId_Null:
			return reflect.Zero(rt), nil

		case TypeId_Array:
			m := v.Value.(*Map).CastSliceMap()

			rv := make(map[string]any, len(m))
			for _, kv := range m {
				if kv.K.TypeId != TypeId_String {
					return reflect.Value{}, ErrInvalidCast
				}
				rv[kv.K.Value.Indirect().(string)] = kv.V.Value.Indirect()
			}
			return reflect.ValueOf(&rv), nil

		default:
			return reflect.Value{}, ErrInvalidCast
		}
	case sliceMapStringAnyRT:
		switch v.TypeId {
		case TypeId_Null:
			return reflect.Zero(rt), nil

		case TypeId_Array:
			m := v.Value.(*Map).CastSliceMap()

			rv := make(generic.SliceMap[string, any], 0, len(m))
			for _, kv := range m {
				if kv.K.TypeId != TypeId_String {
					return reflect.Value{}, ErrInvalidCast
				}
				rv.Add(kv.K.Value.Indirect().(string), kv.V.Value.Indirect())
			}
			return reflect.ValueOf(rv), nil

		default:
			return reflect.Value{}, ErrInvalidCast
		}
	case reflect.PointerTo(sliceMapStringAnyRT):
		switch v.TypeId {
		case TypeId_Null:
			return reflect.Zero(rt), nil

		case TypeId_Array:
			m := v.Value.(*Map).CastSliceMap()

			rv := make(generic.SliceMap[string, any], 0, len(m))
			for _, kv := range m {
				if kv.K.TypeId != TypeId_String {
					return reflect.Value{}, ErrInvalidCast
				}
				rv.Add(kv.K.Value.Indirect().(string), kv.V.Value.Indirect())
			}
			return reflect.ValueOf(&rv), nil

		default:
			return reflect.Value{}, ErrInvalidCast
		}
	case rvRT:
		return v.Reflected.Elem(), nil

	case reflect.PointerTo(rvRT):
		rv := v.Reflected.Elem()
		return reflect.ValueOf(&rv), nil

	case variantRT:
		return reflect.ValueOf(v), nil

	case reflect.PointerTo(variantRT):
		rv := reflect.ValueOf(v)
		return reflect.ValueOf(&rv), nil

	default:
		return reflect.Value{}, ErrInvalidCast
	}
}
