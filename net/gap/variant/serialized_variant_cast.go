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

// CastSerializedVariant 转换已序列化可变类型
func CastSerializedVariant(a any) (ret Variant, err error) {
retry:
	switch v := a.(type) {
	case int:
		return MakeSerializedVariant(Int(v))
	case *int:
		return MakeSerializedVariant((*Int)(v))
	case int8:
		return MakeSerializedVariant(Int8(v))
	case *int8:
		return MakeSerializedVariant((*Int8)(v))
	case int16:
		return MakeSerializedVariant(Int16(v))
	case *int16:
		return MakeSerializedVariant((*Int16)(v))
	case int32:
		return MakeSerializedVariant(Int32(v))
	case *int32:
		return MakeSerializedVariant((*Int32)(v))
	case int64:
		return MakeSerializedVariant(Int64(v))
	case *int64:
		return MakeSerializedVariant((*Int64)(v))
	case uint:
		return MakeSerializedVariant(Uint(v))
	case *uint:
		return MakeSerializedVariant((*Uint)(v))
	case uint8:
		return MakeSerializedVariant(Uint8(v))
	case *uint8:
		return MakeSerializedVariant((*Uint8)(v))
	case uint16:
		return MakeSerializedVariant(Uint16(v))
	case *uint16:
		return MakeSerializedVariant((*Uint16)(v))
	case uint32:
		return MakeSerializedVariant(Uint32(v))
	case *uint32:
		return MakeSerializedVariant((*Uint32)(v))
	case uint64:
		return MakeSerializedVariant(Uint64(v))
	case *uint64:
		return MakeSerializedVariant((*Uint64)(v))
	case float32:
		return MakeSerializedVariant(Float(v))
	case *float32:
		return MakeSerializedVariant((*Float)(v))
	case float64:
		return MakeSerializedVariant(Double(v))
	case *float64:
		return MakeSerializedVariant((*Double)(v))
	case bool:
		return MakeSerializedVariant(Bool(v))
	case *bool:
		return MakeSerializedVariant((*Bool)(v))
	case []byte:
		return MakeSerializedVariant(Bytes(v))
	case *[]byte:
		return MakeSerializedVariant((*Bytes)(v))
	case string:
		return MakeSerializedVariant(String(v))
	case *string:
		return MakeSerializedVariant((*String)(v))
	case uid.Id:
		return MakeSerializedVariant(String(v))
	case *uid.Id:
		return MakeSerializedVariant((*String)(v))
	case nil:
		return MakeSerializedVariant(Null{})
	case Array:
		return MakeSerializedVariant(v)
	case *Array:
		return MakeSerializedVariant(*v)
	case []any:
		arr, err := MakeSerializedArray(v)
		if err != nil {
			return Variant{}, err
		}
		defer func() {
			if !ret.IsValid() {
				arr.Release()
			}
		}()
		return MakeSerializedVariant(arr)
	case *[]any:
		arr, err := MakeSerializedArray(*v)
		if err != nil {
			return Variant{}, err
		}
		defer func() {
			if !ret.IsValid() {
				arr.Release()
			}
		}()
		return MakeSerializedVariant(arr)
	case []reflect.Value:
		arr, err := MakeSerializedArray(v)
		if err != nil {
			return Variant{}, err
		}
		defer func() {
			if !ret.IsValid() {
				arr.Release()
			}
		}()
		return MakeSerializedVariant(arr)
	case *[]reflect.Value:
		arr, err := MakeSerializedArray(*v)
		if err != nil {
			return Variant{}, err
		}
		defer func() {
			if !ret.IsValid() {
				arr.Release()
			}
		}()
		return MakeSerializedVariant(arr)
	case Map:
		return MakeSerializedVariant(v)
	case *Map:
		return MakeSerializedVariant(*v)
	case map[string]any:
		m, err := MakeSerializedMapFromGoMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		defer func() {
			if !ret.IsValid() {
				m.Release()
			}
		}()
		return MakeSerializedVariant(m)
	case *map[string]any:
		m, err := MakeSerializedMapFromGoMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		defer func() {
			if !ret.IsValid() {
				m.Release()
			}
		}()
		return MakeSerializedVariant(m)
	case generic.SliceMap[string, any]:
		m, err := MakeSerializedMapFromSliceMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		defer func() {
			if !ret.IsValid() {
				m.Release()
			}
		}()
		return MakeSerializedVariant(m)
	case *generic.SliceMap[string, any]:
		m, err := MakeSerializedMapFromSliceMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		defer func() {
			if !ret.IsValid() {
				m.Release()
			}
		}()
		return MakeSerializedVariant(m)
	case generic.UnorderedSliceMap[string, any]:
		m, err := MakeSerializedMapFromUnorderedSliceMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		defer func() {
			if !ret.IsValid() {
				m.Release()
			}
		}()
		return MakeSerializedVariant(m)
	case *generic.UnorderedSliceMap[string, any]:
		m, err := MakeSerializedMapFromUnorderedSliceMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		defer func() {
			if !ret.IsValid() {
				m.Release()
			}
		}()
		return MakeSerializedVariant(m)
	case Error:
		return MakeSerializedVariant(&v)
	case *Error:
		return MakeSerializedVariant(v)
	case error:
		return MakeSerializedVariant(MakeError(v))
	case CallChain:
		return MakeSerializedVariant(v)
	case *CallChain:
		return MakeSerializedVariant(*v)
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
		return MakeSerializedVariant(v)
	default:
		return Variant{}, ErrInvalidCast
	}
}
