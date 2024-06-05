package variant

import (
	"git.golaxy.org/core/utils/generic"
	"reflect"
)

var (
	sliceAnyRT          = reflect.TypeFor[[]any]()
	sliceRVRT           = reflect.TypeFor[[]reflect.Value]()
	mapStringAnyRT      = reflect.TypeFor[map[string]any]()
	sliceMapStringAnyRT = reflect.TypeFor[generic.SliceMap[string, any]]()
	rvRT                = reflect.TypeFor[reflect.Value]()
	variantRT           = reflect.TypeFor[Variant]()
)

// CastVariantReflected 转换反射可变类型
func CastVariantReflected(v Variant, rt reflect.Type) (reflect.Value, error) {
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
				rv = append(rv, reflect.ValueOf(it.Value.Indirect()))
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
				rv = append(rv, reflect.ValueOf(it.Value.Indirect()))
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
