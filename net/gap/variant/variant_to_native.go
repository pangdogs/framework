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
	"fmt"
	"reflect"

	"git.golaxy.org/core/utils/generic"
)

var (
	// ErrInvalidCast 表示动态值无法安全转换为目标 Go 类型。
	ErrInvalidCast = fmt.Errorf("%w: invalid cast", ErrVariant)
)

var (
	sliceAnyRT                   = reflect.TypeFor[[]any]()
	sliceRVRT                    = reflect.TypeFor[[]reflect.Value]()
	mapStringAnyRT               = reflect.TypeFor[map[string]any]()
	sliceMapStringAnyRT          = reflect.TypeFor[generic.SliceMap[string, any]]()
	unorderedSliceMapStringAnyRT = reflect.TypeFor[generic.UnorderedSliceMap[string, any]]()
	rvRT                         = reflect.TypeFor[reflect.Value]()
	variantRT                    = reflect.TypeFor[Variant]()
	readableValueRT              = reflect.TypeFor[ReadableValue]()
	valueInterfaceRT             = reflect.TypeFor[Value]()
)

// ToNative 将动态值转换为 valueRT；数值转换不允许缩窄底层存储宽度。
func (v Variant) ToNative(valueRT reflect.Type) (reflect.Value, error) {
	if !v.IsValid() {
		return reflect.Value{}, ErrInvalidCast
	}

	{
		retRV := v.Reflected
		if !retRV.IsValid() {
			retRV = reflect.ValueOf(v.Value)
		}
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

		if (valueRT == readableValueRT || valueRT == valueInterfaceRT) && reflect.PointerTo(retRT).Implements(valueRT) {
			ptrRV := reflect.New(retRT)
			ptrRV.Elem().Set(retRV)
			return ptrRV, nil
		}

		if retRT.Kind() == reflect.Pointer {
			retRV = retRV.Elem()
			retRT = retRV.Type()
			goto retry
		}
	}

	switch valueRT.Kind() {
	case reflect.Interface, reflect.Array, reflect.Map, reflect.Pointer, reflect.Slice:
		if v.TypeID == TypeID_Null {
			return reflect.Zero(valueRT), nil
		}
	case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func:
		return reflect.Value{}, ErrInvalidCast
	}

	switch valueRT {
	case sliceAnyRT, reflect.PointerTo(sliceAnyRT):
		switch v.TypeID {
		case TypeID_Array:
			arr, ok := indirectArray(v.Value)
			if !ok {
				return reflect.Value{}, ErrInvalidCast
			}

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
		switch v.TypeID {
		case TypeID_Array:
			arr, ok := indirectArray(v.Value)
			if !ok {
				return reflect.Value{}, ErrInvalidCast
			}

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
		switch v.TypeID {
		case TypeID_Map:
			m, ok := indirectMap(v.Value)
			if !ok {
				return reflect.Value{}, ErrInvalidCast
			}

			rv := make(map[string]any, len(m))
			for _, kv := range m {
				if kv.K.TypeID != TypeID_String {
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
		switch v.TypeID {
		case TypeID_Map:
			m, ok := indirectMap(v.Value)
			if !ok {
				return reflect.Value{}, ErrInvalidCast
			}

			rv := make(generic.SliceMap[string, any], 0, len(m))
			for _, kv := range m {
				if kv.K.TypeID != TypeID_String {
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
		switch v.TypeID {
		case TypeID_Map:
			m, ok := indirectMap(v.Value)
			if !ok {
				return reflect.Value{}, ErrInvalidCast
			}

			rv := make(generic.UnorderedSliceMap[string, any], 0, len(m))
			for _, kv := range m {
				if kv.K.TypeID != TypeID_String {
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

	case rvRT, reflect.PointerTo(rvRT):
		rv := v.Reflected
		if !rv.IsValid() {
			rv = reflect.ValueOf(v.Value)
		}
		if rv.Kind() == reflect.Pointer {
			rv = rv.Elem()
		}

		if valueRT.Kind() == reflect.Pointer {
			return reflect.ValueOf(&rv), nil
		} else {
			return rv, nil
		}

	case variantRT, reflect.PointerTo(variantRT):
		if valueRT.Kind() == reflect.Pointer {
			return reflect.ValueOf(&v), nil
		} else {
			return reflect.ValueOf(v), nil
		}
	}

	switch v.TypeID {
	case TypeID_Array:
		switch valueRT.Kind() {
		case reflect.Array, reflect.Slice:
			return convertArrayTo(v, valueRT)
		}

	case TypeID_Map:
		if valueRT.Kind() == reflect.Map {
			return convertMapTo(v, valueRT)
		}
	}

	return reflect.Value{}, ErrInvalidCast
}

func convertArrayTo(v Variant, valueRT reflect.Type) (reflect.Value, error) {
	arr, ok := indirectArray(v.Value)
	if !ok {
		return reflect.Value{}, ErrInvalidCast
	}

	var retRV reflect.Value
	switch valueRT.Kind() {
	case reflect.Array:
		if valueRT.Len() != len(arr) {
			return reflect.Value{}, ErrInvalidCast
		}
		retRV = reflect.New(valueRT).Elem()

	case reflect.Slice:
		retRV = reflect.MakeSlice(valueRT, len(arr), len(arr))

	default:
		return reflect.Value{}, ErrInvalidCast
	}

	elemRT := valueRT.Elem()
	for i := range arr {
		elemRV, err := arr[i].ToNative(elemRT)
		if err != nil {
			return reflect.Value{}, ErrInvalidCast
		}
		elemRV, err = assignableOrConvert(elemRV, elemRT)
		if err != nil {
			return reflect.Value{}, err
		}
		retRV.Index(i).Set(elemRV)
	}

	return retRV, nil
}

func convertMapTo(v Variant, valueRT reflect.Type) (reflect.Value, error) {
	m, ok := indirectMap(v.Value)
	if !ok {
		return reflect.Value{}, ErrInvalidCast
	}

	keyRT := valueRT.Key()
	valueElemRT := valueRT.Elem()
	retRV := reflect.MakeMapWithSize(valueRT, len(m))

	for _, kv := range m {
		keyRV, err := kv.K.ToNative(keyRT)
		if err != nil {
			return reflect.Value{}, ErrInvalidCast
		}
		keyRV, err = assignableOrConvert(keyRV, keyRT)
		if err != nil {
			return reflect.Value{}, err
		}

		valueRV, err := kv.V.ToNative(valueElemRT)
		if err != nil {
			return reflect.Value{}, ErrInvalidCast
		}
		valueRV, err = assignableOrConvert(valueRV, valueElemRT)
		if err != nil {
			return reflect.Value{}, err
		}

		retRV.SetMapIndex(keyRV, valueRV)
	}

	return retRV, nil
}

func assignableOrConvert(v reflect.Value, rt reflect.Type) (reflect.Value, error) {
	if v.Type().AssignableTo(rt) {
		return v, nil
	}
	if v.CanConvert(rt) {
		return v.Convert(rt), nil
	}
	return reflect.Value{}, ErrInvalidCast
}

func indirectArray(v ReadableValue) ([]Variant, bool) {
	switch arr := v.(type) {
	case Array:
		if arr.IsSnapshot {
			return nil, false
		}
		return arr.Items, true
	case *Array:
		if arr.IsSnapshot {
			return nil, false
		}
		return arr.Items, true
	default:
		return nil, false
	}
}

func indirectMap(v ReadableValue) (generic.UnorderedSliceMap[Variant, Variant], bool) {
	switch m := v.(type) {
	case Map:
		return m.Entries, true
	case *Map:
		return m.Entries, true
	default:
		return nil, false
	}
}
