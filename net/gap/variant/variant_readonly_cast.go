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

// CastReadonlyVariant 转换只读可变类型
func CastReadonlyVariant(a any) (Variant, error) {
retry:
	switch v := a.(type) {
	case int:
		return MakeReadonlyVariant(Int(v))
	case *int:
		return MakeReadonlyVariant((*Int)(v))
	case int8:
		return MakeReadonlyVariant(Int8(v))
	case *int8:
		return MakeReadonlyVariant((*Int8)(v))
	case int16:
		return MakeReadonlyVariant(Int16(v))
	case *int16:
		return MakeReadonlyVariant((*Int16)(v))
	case int32:
		return MakeReadonlyVariant(Int32(v))
	case *int32:
		return MakeReadonlyVariant((*Int32)(v))
	case int64:
		return MakeReadonlyVariant(Int64(v))
	case *int64:
		return MakeReadonlyVariant((*Int64)(v))
	case uint:
		return MakeReadonlyVariant(Uint(v))
	case *uint:
		return MakeReadonlyVariant((*Uint)(v))
	case uint8:
		return MakeReadonlyVariant(Uint8(v))
	case *uint8:
		return MakeReadonlyVariant((*Uint8)(v))
	case uint16:
		return MakeReadonlyVariant(Uint16(v))
	case *uint16:
		return MakeReadonlyVariant((*Uint16)(v))
	case uint32:
		return MakeReadonlyVariant(Uint32(v))
	case *uint32:
		return MakeReadonlyVariant((*Uint32)(v))
	case uint64:
		return MakeReadonlyVariant(Uint64(v))
	case *uint64:
		return MakeReadonlyVariant((*Uint64)(v))
	case float32:
		return MakeReadonlyVariant(Float(v))
	case *float32:
		return MakeReadonlyVariant((*Float)(v))
	case float64:
		return MakeReadonlyVariant(Double(v))
	case *float64:
		return MakeReadonlyVariant((*Double)(v))
	case bool:
		return MakeReadonlyVariant(Bool(v))
	case *bool:
		return MakeReadonlyVariant((*Bool)(v))
	case []byte:
		return MakeReadonlyVariant(Bytes(v))
	case *[]byte:
		return MakeReadonlyVariant((*Bytes)(v))
	case string:
		return MakeReadonlyVariant(String(v))
	case *string:
		return MakeReadonlyVariant((*String)(v))
	case uid.Id:
		return MakeReadonlyVariant(String(v))
	case *uid.Id:
		return MakeReadonlyVariant((*String)(v))
	case nil:
		return MakeReadonlyVariant(Null{})
	case Array:
		return MakeReadonlyVariant(v)
	case *Array:
		return MakeReadonlyVariant(*v)
	case []any:
		arr, err := MakeReadonlyArray(v)
		if err != nil {
			return Variant{}, err
		}
		return MakeReadonlyVariant(arr)
	case *[]any:
		arr, err := MakeReadonlyArray(*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeReadonlyVariant(arr)
	case []reflect.Value:
		arr, err := MakeReadonlyArray(v)
		if err != nil {
			return Variant{}, err
		}
		return MakeReadonlyVariant(arr)
	case *[]reflect.Value:
		arr, err := MakeReadonlyArray(*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeReadonlyVariant(arr)
	case Map:
		return MakeReadonlyVariant(v)
	case *Map:
		return MakeReadonlyVariant(*v)
	case map[string]any:
		m, err := MakeReadonlyMapFromGoMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		return MakeReadonlyVariant(m)
	case *map[string]any:
		m, err := MakeReadonlyMapFromGoMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeReadonlyVariant(m)
	case generic.SliceMap[string, any]:
		m, err := MakeReadonlyMapFromSliceMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		return MakeReadonlyVariant(m)
	case *generic.SliceMap[string, any]:
		m, err := MakeReadonlyMapFromSliceMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeReadonlyVariant(m)
	case generic.UnorderedSliceMap[string, any]:
		m, err := MakeReadonlyMapFromUnorderedSliceMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		return MakeReadonlyVariant(m)
	case *generic.UnorderedSliceMap[string, any]:
		m, err := MakeReadonlyMapFromUnorderedSliceMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeReadonlyVariant(m)
	case Error:
		return MakeReadonlyVariant(&v)
	case *Error:
		return MakeReadonlyVariant(v)
	case error:
		return MakeReadonlyVariant(MakeError(v))
	case CallChain:
		return MakeReadonlyVariant(v)
	case *CallChain:
		return MakeReadonlyVariant(*v)
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
		return MakeReadonlyVariant(v)
	default:
		return Variant{}, ErrInvalidCast
	}
}
