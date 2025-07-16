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
	"git.golaxy.org/core/utils/uid"
	"reflect"
)

// CastVariant 转换只读可变类型
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
		return MakeVariant(*v)
	case *any:
		a = *v
		goto retry
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
		return MakeVariant(*v)
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
	case generic.UnorderedSliceMap[string, any]:
		m, err := MakeMapFromUnorderedSliceMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(m)
	case *generic.UnorderedSliceMap[string, any]:
		m, err := MakeMapFromUnorderedSliceMap[string, any](*v)
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
	case ReadableValue:
		return MakeVariant(v)
	default:
		return Variant{}, ErrInvalidCast
	}
}
