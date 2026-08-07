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
	"unsafe"

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/uid"
)

// ToVariant 将受支持的 Go 值或 reflect.Value 转换为 GAP 动态值。
func ToVariant(a any) (Variant, error) {
retry:
	switch v := a.(type) {
	case nil:
		return NewVariant(Null{})
	case int:
		return NewVariant(Int(v))
	case *int:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*Int)(v))
	case int8:
		return NewVariant(Int8(v))
	case *int8:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*Int8)(v))
	case int16:
		return NewVariant(Int16(v))
	case *int16:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*Int16)(v))
	case int32:
		return NewVariant(Int32(v))
	case *int32:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*Int32)(v))
	case int64:
		return NewVariant(Int64(v))
	case *int64:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*Int64)(v))
	case uint:
		return NewVariant(Uint(v))
	case *uint:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*Uint)(v))
	case uint8:
		return NewVariant(Uint8(v))
	case *uint8:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*Uint8)(v))
	case uint16:
		return NewVariant(Uint16(v))
	case *uint16:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*Uint16)(v))
	case uint32:
		return NewVariant(Uint32(v))
	case *uint32:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*Uint32)(v))
	case uint64:
		return NewVariant(Uint64(v))
	case *uint64:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*Uint64)(v))
	case float32:
		return NewVariant(Float(v))
	case *float32:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*Float)(v))
	case float64:
		return NewVariant(Double(v))
	case *float64:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*Double)(v))
	case bool:
		return NewVariant(Bool(v))
	case *bool:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*Bool)(v))
	case []byte:
		return NewVariant(Bytes(v))
	case *[]byte:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*Bytes)(v))
	case string:
		return NewVariant(String(v))
	case *string:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*String)(v))
	case uid.Id:
		return NewVariant(String(v))
	case *uid.Id:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant((*String)(v))
	case Array:
		return NewVariant(v)
	case *Array:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant(*v)
	case *any:
		if v == nil {
			a = nil
			goto retry
		}
		a = *v
		goto retry
	case []any:
		arr, err := NewArray(v)
		if err != nil {
			return Variant{}, err
		}
		return NewVariant(arr)
	case *[]any:
		if v == nil {
			a = nil
			goto retry
		}
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
		if v == nil {
			a = nil
			goto retry
		}
		arr, err := NewArray(*v)
		if err != nil {
			return Variant{}, err
		}
		return NewVariant(arr)
	case Map:
		return NewVariant(v)
	case *Map:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant(*v)
	case map[string]any:
		m, err := NewMapFromGoMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		return NewVariant(m)
	case *map[string]any:
		if v == nil {
			a = nil
			goto retry
		}
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
		if v == nil {
			a = nil
			goto retry
		}
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
		if v == nil {
			a = nil
			goto retry
		}
		m, err := NewMapFromUnorderedSliceMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		return NewVariant(m)
	case Error:
		return NewVariant(&v)
	case *Error:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant(v)
	case error:
		return NewVariant(NewError(v))
	case CallChain:
		return NewVariant(v)
	case *CallChain:
		if v == nil {
			a = nil
			goto retry
		}
		return NewVariant(*v)
	case reflect.Value:
		if !v.CanInterface() {
			return Variant{}, ErrInvalidCast
		}
		a = v.Interface()
		goto retry
	case *reflect.Value:
		if v == nil {
			a = nil
			goto retry
		}
		if !v.CanInterface() {
			return Variant{}, ErrInvalidCast
		}
		a = v.Interface()
		goto retry
	case Variant:
		return v, nil
	case *Variant:
		if v == nil {
			a = nil
			goto retry
		}
		return *v, nil
	case ReadableValue:
		if (*[2]unsafe.Pointer)(unsafe.Pointer(&v))[1] == nil {
			a = nil
			goto retry
		}
		return NewVariant(v)
	default:
		rv := reflect.ValueOf(a)
		for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
			if rv.IsNil() {
				a = nil
				goto retry
			}
			rv = rv.Elem()
		}

		switch rv.Kind() {
		case reflect.Array, reflect.Slice:
			arr, err := newArrayFromNativeValue(rv)
			if err != nil {
				return Variant{}, err
			}
			return NewVariant(arr)

		case reflect.Map:
			m, err := newMapFromNativeValue(rv)
			if err != nil {
				return Variant{}, err
			}
			return NewVariant(m)
		}

		return Variant{}, ErrInvalidCast
	}
}

func newArrayFromNativeValue(rv reflect.Value) (Array, error) {
	ret := Array{
		Items: make([]Variant, 0, rv.Len()),
	}

	for i := 0; i < rv.Len(); i++ {
		item, err := toVariantFromNativeValue(rv.Index(i))
		if err != nil {
			return Array{}, err
		}
		ret.Items = append(ret.Items, item)
	}

	return ret, nil
}

func newMapFromNativeValue(rv reflect.Value) (Map, error) {
	ret := Map{
		Entries: make(generic.UnorderedSliceMap[Variant, Variant], 0, rv.Len()),
	}

	iter := rv.MapRange()
	for iter.Next() {
		k, err := toVariantFromNativeValue(iter.Key())
		if err != nil {
			return Map{}, err
		}

		v, err := toVariantFromNativeValue(iter.Value())
		if err != nil {
			return Map{}, err
		}

		ret.Entries = append(ret.Entries, generic.UnorderedKV[Variant, Variant]{K: k, V: v})
	}

	return ret, nil
}

func toVariantFromNativeValue(rv reflect.Value) (Variant, error) {
	if !rv.IsValid() {
		return ToVariant(nil)
	}
	for rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return ToVariant(nil)
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Pointer && rv.CanAddr() {
		addr := rv.Addr()
		if addr.CanInterface() {
			return ToVariant(addr.Interface())
		}
	}
	if rv.CanInterface() {
		return ToVariant(rv.Interface())
	}
	return Variant{}, ErrInvalidCast
}
