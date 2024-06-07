package variant

import (
	"git.golaxy.org/core/utils/generic"
	"reflect"
)

var (
	sliceAnyRT                   = reflect.TypeFor[[]any]()
	sliceRVRT                    = reflect.TypeFor[[]reflect.Value]()
	mapStringAnyRT               = reflect.TypeFor[map[string]any]()
	sliceMapStringAnyRT          = reflect.TypeFor[generic.SliceMap[string, any]]()
	unorderedSliceMapStringAnyRT = reflect.TypeFor[generic.UnorderedSliceMap[string, any]]()
	rvRT                         = reflect.TypeFor[reflect.Value]()
	variantRT                    = reflect.TypeFor[Variant]()
)

// CastVariantReflected 转换反射可变类型
func CastVariantReflected(variant Variant, valueRT reflect.Type) (reflect.Value, error) {
	if !variant.Reflected.IsValid() {
		return reflect.Value{}, ErrInvalidCast
	}

	switch valueRT.Kind() {
	case reflect.Array, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if variant.TypeId == TypeId_Null {
			return reflect.Zero(valueRT), nil
		}
	case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func:
		return reflect.Value{}, ErrInvalidCast
	}

	switch valueRT {
	case sliceAnyRT, reflect.PointerTo(sliceAnyRT):
		switch variant.TypeId {
		case TypeId_Array:
			arr := *variant.Value.(*Array)

			rv := make([]any, 0, len(arr))
			for _, it := range arr {
				rv = append(rv, it.Value.Indirect())
			}

			if valueRT.Kind() == reflect.Pointer {
				return reflect.ValueOf(&rv), nil
			} else {
				return reflect.ValueOf(rv), nil
			}
		}

	case sliceRVRT, reflect.PointerTo(sliceRVRT):
		switch variant.TypeId {
		case TypeId_Array:
			arr := *variant.Value.(*Array)

			rv := make([]reflect.Value, 0, len(arr))
			for _, it := range arr {
				rv = append(rv, reflect.ValueOf(it.Value.Indirect()))
			}

			if valueRT.Kind() == reflect.Pointer {
				return reflect.ValueOf(&rv), nil
			} else {
				return reflect.ValueOf(rv), nil
			}
		}

	case mapStringAnyRT, reflect.PointerTo(mapStringAnyRT):
		switch variant.TypeId {
		case TypeId_Map:
			m := *variant.Value.(*Map).CastUnorderedSliceMap()

			rv := make(map[string]any, len(m))
			for _, kv := range m {
				if kv.K.TypeId != TypeId_String {
					return reflect.Value{}, ErrInvalidCast
				}
				rv[kv.K.Value.Indirect().(string)] = kv.V.Value.Indirect()
			}

			if valueRT.Kind() == reflect.Pointer {
				return reflect.ValueOf(&rv), nil
			} else {
				return reflect.ValueOf(rv), nil
			}
		}

	case sliceMapStringAnyRT, reflect.PointerTo(sliceMapStringAnyRT):
		switch variant.TypeId {
		case TypeId_Map:
			m := *variant.Value.(*Map).CastUnorderedSliceMap()

			rv := make(generic.SliceMap[string, any], 0, len(m))
			for _, kv := range m {
				if kv.K.TypeId != TypeId_String {
					return reflect.Value{}, ErrInvalidCast
				}
				rv.Add(kv.K.Value.Indirect().(string), kv.V.Value.Indirect())
			}

			if valueRT.Kind() == reflect.Pointer {
				return reflect.ValueOf(&rv), nil
			} else {
				return reflect.ValueOf(rv), nil
			}
		}

	case unorderedSliceMapStringAnyRT, reflect.PointerTo(unorderedSliceMapStringAnyRT):
		switch variant.TypeId {
		case TypeId_Map:
			m := variant.Value.(*Map).CastUnorderedSliceMap()

			for _, kv := range *m {
				if kv.K.TypeId != TypeId_String {
					return reflect.Value{}, ErrInvalidCast
				}
			}

			if valueRT.Kind() == reflect.Pointer {
				return reflect.ValueOf(m), nil
			} else {
				return reflect.ValueOf(*m), nil
			}
		}

	case rvRT, reflect.PointerTo(rvRT):
		rv := variant.Reflected.Elem()

		if valueRT.Kind() == reflect.Pointer {
			return reflect.ValueOf(&rv), nil
		} else {
			return rv, nil
		}

	case variantRT, reflect.PointerTo(variantRT):
		rv := reflect.ValueOf(variant)

		if valueRT.Kind() == reflect.Pointer {
			return reflect.ValueOf(&rv), nil
		} else {
			return rv, nil
		}
	}

	return reflect.Value{}, ErrInvalidCast
}
