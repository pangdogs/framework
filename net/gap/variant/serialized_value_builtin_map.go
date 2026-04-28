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
	"cmp"

	"git.golaxy.org/core/utils/generic"
)

// NewSerializedMapFromGoMap 创建已序列化map
func NewSerializedMapFromGoMap[K comparable, V any](m map[K]V) (ret SerializedMap, err error) {
	ret.Map = make(Map, 0, len(m))
	ret.entries = make([]serializedMapEntry, 0, len(m))
	defer func() {
		if err != nil {
			ret.Release()
			ret = SerializedMap{}
		}
	}()

	for k, v := range m {
		varK, err := CastSerializedVariant(k)
		if err != nil {
			return SerializedMap{}, err
		}

		varV, err := CastSerializedVariant(v)
		if err != nil {
			return SerializedMap{}, err
		}

		ret.Map.ToUnorderedSliceMap().Add(varK.Ref(), varV.Ref())
		ret.entries = append(ret.entries, serializedMapEntry{K: varK, V: varV})
	}

	return ret, nil
}

// NewSerializedMapFromSliceMap 创建已序列化map
func NewSerializedMapFromSliceMap[K cmp.Ordered, V any](m generic.SliceMap[K, V]) (ret SerializedMap, err error) {
	ret.Map = make(Map, 0, len(m))
	ret.entries = make([]serializedMapEntry, 0, len(m))
	defer func() {
		if err != nil {
			ret.Release()
			ret = SerializedMap{}
		}
	}()

	for i := range m {
		kv := &m[i]

		varK, err := CastSerializedVariant(&kv.K)
		if err != nil {
			return SerializedMap{}, err
		}

		varV, err := CastSerializedVariant(&kv.V)
		if err != nil {
			return SerializedMap{}, err
		}

		ret.Map.ToUnorderedSliceMap().Add(varK.Ref(), varV.Ref())
		ret.entries = append(ret.entries, serializedMapEntry{K: varK, V: varV})
	}

	return ret, nil
}

// NewSerializedMapFromUnorderedSliceMap 创建已序列化map
func NewSerializedMapFromUnorderedSliceMap[K comparable, V any](m generic.UnorderedSliceMap[K, V]) (ret SerializedMap, err error) {
	ret.Map = make(Map, 0, len(m))
	ret.entries = make([]serializedMapEntry, 0, len(m))
	defer func() {
		if err != nil {
			ret.Release()
			ret = SerializedMap{}
		}
	}()

	for i := range m {
		kv := &m[i]

		varK, err := CastSerializedVariant(&kv.K)
		if err != nil {
			return SerializedMap{}, err
		}

		varV, err := CastSerializedVariant(&kv.V)
		if err != nil {
			return SerializedMap{}, err
		}

		ret.Map.ToUnorderedSliceMap().Add(varK.Ref(), varV.Ref())
		ret.entries = append(ret.entries, serializedMapEntry{K: varK, V: varV})
	}

	return ret, nil
}

type serializedMapEntry struct {
	K SerializedVariant
	V SerializedVariant
}

// SerializedMap 已序列化map
type SerializedMap struct {
	Map
	entries []serializedMapEntry
}

// Release 释放缓存，缓存释放后请勿再使用或引用变体值
func (m SerializedMap) Release() {
	for i := range m.entries {
		m.entries[i].K.Release()
		m.entries[i].V.Release()
	}
}

// Ref 引用变体值
func (m SerializedMap) Ref() Map {
	return m.Map
}
