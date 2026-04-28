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
	"reflect"

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/uid"
)

// CastSerializedVariant 转换已序列化变体
func CastSerializedVariant(a any) (ret SerializedVariant, err error) {
retry:
	switch v := a.(type) {
	case int:
		return NewSerializedVariant(Int(v))
	case *int:
		return NewSerializedVariant((*Int)(v))
	case int8:
		return NewSerializedVariant(Int8(v))
	case *int8:
		return NewSerializedVariant((*Int8)(v))
	case int16:
		return NewSerializedVariant(Int16(v))
	case *int16:
		return NewSerializedVariant((*Int16)(v))
	case int32:
		return NewSerializedVariant(Int32(v))
	case *int32:
		return NewSerializedVariant((*Int32)(v))
	case int64:
		return NewSerializedVariant(Int64(v))
	case *int64:
		return NewSerializedVariant((*Int64)(v))
	case uint:
		return NewSerializedVariant(Uint(v))
	case *uint:
		return NewSerializedVariant((*Uint)(v))
	case uint8:
		return NewSerializedVariant(Uint8(v))
	case *uint8:
		return NewSerializedVariant((*Uint8)(v))
	case uint16:
		return NewSerializedVariant(Uint16(v))
	case *uint16:
		return NewSerializedVariant((*Uint16)(v))
	case uint32:
		return NewSerializedVariant(Uint32(v))
	case *uint32:
		return NewSerializedVariant((*Uint32)(v))
	case uint64:
		return NewSerializedVariant(Uint64(v))
	case *uint64:
		return NewSerializedVariant((*Uint64)(v))
	case float32:
		return NewSerializedVariant(Float(v))
	case *float32:
		return NewSerializedVariant((*Float)(v))
	case float64:
		return NewSerializedVariant(Double(v))
	case *float64:
		return NewSerializedVariant((*Double)(v))
	case bool:
		return NewSerializedVariant(Bool(v))
	case *bool:
		return NewSerializedVariant((*Bool)(v))
	case []byte:
		return NewSerializedVariant(Bytes(v))
	case *[]byte:
		return NewSerializedVariant((*Bytes)(v))
	case string:
		return NewSerializedVariant(String(v))
	case *string:
		return NewSerializedVariant((*String)(v))
	case uid.Id:
		return NewSerializedVariant(String(v))
	case *uid.Id:
		return NewSerializedVariant((*String)(v))
	case nil:
		return NewSerializedVariant(Null{})
	case Array:
		return NewSerializedVariant(v)
	case *Array:
		return NewSerializedVariant(*v)
	case *any:
		a = *v
		goto retry
	case []any:
		arr, err := NewSerializedArray(v)
		if err != nil {
			return SerializedVariant{}, err
		}
		return wrappedSerializedVariant(arr, arr), nil
	case *[]any:
		arr, err := NewSerializedArray(*v)
		if err != nil {
			return SerializedVariant{}, err
		}
		return wrappedSerializedVariant(arr, arr), nil
	case []reflect.Value:
		arr, err := NewSerializedArray(v)
		if err != nil {
			return SerializedVariant{}, err
		}
		return wrappedSerializedVariant(arr, arr), nil
	case *[]reflect.Value:
		arr, err := NewSerializedArray(*v)
		if err != nil {
			return SerializedVariant{}, err
		}
		return wrappedSerializedVariant(arr, arr), nil
	case Map:
		return NewSerializedVariant(v)
	case *Map:
		return NewSerializedVariant(*v)
	case map[string]any:
		m, err := NewSerializedMapFromGoMap[string, any](v)
		if err != nil {
			return SerializedVariant{}, err
		}
		return wrappedSerializedVariant(m, m), nil
	case *map[string]any:
		m, err := NewSerializedMapFromGoMap[string, any](*v)
		if err != nil {
			return SerializedVariant{}, err
		}
		return wrappedSerializedVariant(m, m), nil
	case generic.SliceMap[string, any]:
		m, err := NewSerializedMapFromSliceMap[string, any](v)
		if err != nil {
			return SerializedVariant{}, err
		}
		return wrappedSerializedVariant(m, m), nil
	case *generic.SliceMap[string, any]:
		m, err := NewSerializedMapFromSliceMap[string, any](*v)
		if err != nil {
			return SerializedVariant{}, err
		}
		return wrappedSerializedVariant(m, m), nil
	case generic.UnorderedSliceMap[string, any]:
		m, err := NewSerializedMapFromUnorderedSliceMap[string, any](v)
		if err != nil {
			return SerializedVariant{}, err
		}
		return wrappedSerializedVariant(m, m), nil
	case *generic.UnorderedSliceMap[string, any]:
		m, err := NewSerializedMapFromUnorderedSliceMap[string, any](*v)
		if err != nil {
			return SerializedVariant{}, err
		}
		return wrappedSerializedVariant(m, m), nil
	case SerializedArray:
		return wrappedSerializedVariant(v, v), nil
	case *SerializedArray:
		return wrappedSerializedVariant(*v, *v), nil
	case SerializedMap:
		return wrappedSerializedVariant(v, v), nil
	case *SerializedMap:
		return wrappedSerializedVariant(*v, *v), nil
	case Error:
		return NewSerializedVariant(&v)
	case *Error:
		return NewSerializedVariant(v)
	case error:
		return NewSerializedVariant(NewError(v))
	case CallChain:
		return NewSerializedVariant(v)
	case *CallChain:
		return NewSerializedVariant(*v)
	case reflect.Value:
		if !v.CanInterface() {
			return SerializedVariant{}, ErrInvalidCast
		}
		a = v.Interface()
		goto retry
	case *reflect.Value:
		if !v.CanInterface() {
			return SerializedVariant{}, ErrInvalidCast
		}
		a = v.Interface()
		goto retry
	case Variant:
		return NewSerializedVariant(v.Value)
	case *Variant:
		return NewSerializedVariant(v.Value)
	case SerializedVariant:
		return v, nil
	case *SerializedVariant:
		return *v, nil
	case ReadableValue:
		return NewSerializedVariant(v)
	default:
		return SerializedVariant{}, ErrInvalidCast
	}
}
