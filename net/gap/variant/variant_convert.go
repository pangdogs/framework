/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

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

func (v Variant) Convert(valueRT reflect.Type) (reflect.Value, error) {
	if !v.Reflected.IsValid() {
		return reflect.Value{}, ErrInvalidCast
	}

	{
		retRV := v.Reflected
		retRT := retRV.Type()

	retry:
		if retRT.AssignableTo(valueRT) {
			return retRV, nil
		}

		if retRV.CanConvert(valueRT) {
			if retRT.Size() > valueRT.Size() {
				return reflect.Value{}, ErrInvalidCast
			}
			return retRV.Convert(valueRT), nil
		}

		if retRT.Kind() == reflect.Pointer {
			retRV = retRV.Elem()
			retRT = retRV.Type()
			goto retry
		}
	}

	switch valueRT.Kind() {
	case reflect.Array, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if v.TypeId == TypeId_Null {
			return reflect.Zero(valueRT), nil
		}
	case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func:
		return reflect.Value{}, ErrInvalidCast
	}

	switch valueRT {
	case sliceAnyRT, reflect.PointerTo(sliceAnyRT):
		switch v.TypeId {
		case TypeId_Array:
			arr := *v.Value.(*Array)

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
		switch v.TypeId {
		case TypeId_Array:
			arr := *v.Value.(*Array)

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
		switch v.TypeId {
		case TypeId_Map:
			m := *v.Value.(*Map).ToUnorderedSliceMap()

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
		switch v.TypeId {
		case TypeId_Map:
			m := *v.Value.(*Map).ToUnorderedSliceMap()

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
		switch v.TypeId {
		case TypeId_Map:
			m := v.Value.(*Map).ToUnorderedSliceMap()

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
		rv := v.Reflected.Elem()

		if valueRT.Kind() == reflect.Pointer {
			return reflect.ValueOf(&rv), nil
		} else {
			return rv, nil
		}

	case variantRT, reflect.PointerTo(variantRT):
		rv := reflect.ValueOf(v)

		if valueRT.Kind() == reflect.Pointer {
			return reflect.ValueOf(&rv), nil
		} else {
			return rv, nil
		}
	}

	return reflect.Value{}, ErrInvalidCast
}
