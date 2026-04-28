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

// CastVariant 转换为变体
func CastVariant(a any) (Variant, error) {
retry:
	switch v := a.(type) {
	case int:
		return NewVariant(Int(v))
	case *int:
		return NewVariant((*Int)(v))
	case int8:
		return NewVariant(Int8(v))
	case *int8:
		return NewVariant((*Int8)(v))
	case int16:
		return NewVariant(Int16(v))
	case *int16:
		return NewVariant((*Int16)(v))
	case int32:
		return NewVariant(Int32(v))
	case *int32:
		return NewVariant((*Int32)(v))
	case int64:
		return NewVariant(Int64(v))
	case *int64:
		return NewVariant((*Int64)(v))
	case uint:
		return NewVariant(Uint(v))
	case *uint:
		return NewVariant((*Uint)(v))
	case uint8:
		return NewVariant(Uint8(v))
	case *uint8:
		return NewVariant((*Uint8)(v))
	case uint16:
		return NewVariant(Uint16(v))
	case *uint16:
		return NewVariant((*Uint16)(v))
	case uint32:
		return NewVariant(Uint32(v))
	case *uint32:
		return NewVariant((*Uint32)(v))
	case uint64:
		return NewVariant(Uint64(v))
	case *uint64:
		return NewVariant((*Uint64)(v))
	case float32:
		return NewVariant(Float(v))
	case *float32:
		return NewVariant((*Float)(v))
	case float64:
		return NewVariant(Double(v))
	case *float64:
		return NewVariant((*Double)(v))
	case bool:
		return NewVariant(Bool(v))
	case *bool:
		return NewVariant((*Bool)(v))
	case []byte:
		return NewVariant(Bytes(v))
	case *[]byte:
		return NewVariant((*Bytes)(v))
	case string:
		return NewVariant(String(v))
	case *string:
		return NewVariant((*String)(v))
	case uid.Id:
		return NewVariant(String(v))
	case *uid.Id:
		return NewVariant((*String)(v))
	case nil:
		return NewVariant(Null{})
	case Array:
		return NewVariant(v)
	case *Array:
		return NewVariant(*v)
	case *any:
		a = *v
		goto retry
	case []any:
		arr, err := NewArray(v)
		if err != nil {
			return Variant{}, err
		}
		return NewVariant(arr)
	case *[]any:
		arr, err := NewArray(*v)
		if err != nil {
			return Variant{}, err
		}
		return NewVariant(arr)
	case []reflect.Value:
		arr, err := NewArray(v)
		if err != nil {
			return Variant{}, err
		}
		return NewVariant(arr)
	case *[]reflect.Value:
		arr, err := NewArray(*v)
		if err != nil {
			return Variant{}, err
		}
		return NewVariant(arr)
	case Map:
		return NewVariant(v)
	case *Map:
		return NewVariant(*v)
	case map[string]any:
		m, err := NewMapFromGoMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		return NewVariant(m)
	case *map[string]any:
		m, err := NewMapFromGoMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		return NewVariant(m)
	case generic.SliceMap[string, any]:
		m, err := NewMapFromSliceMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		return NewVariant(m)
	case *generic.SliceMap[string, any]:
		m, err := NewMapFromSliceMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		return NewVariant(m)
	case generic.UnorderedSliceMap[string, any]:
		m, err := NewMapFromUnorderedSliceMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		return NewVariant(m)
	case *generic.UnorderedSliceMap[string, any]:
		m, err := NewMapFromUnorderedSliceMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		return NewVariant(m)
	case Error:
		return NewVariant(&v)
	case *Error:
		return NewVariant(v)
	case error:
		return NewVariant(NewError(v))
	case CallChain:
		return NewVariant(v)
	case *CallChain:
		return NewVariant(*v)
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
	case ReadableValue:
		return NewVariant(v)
	default:
		return Variant{}, ErrInvalidCast
	}
}
