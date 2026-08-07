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
	"io"

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/utils/binaryutil"
)

// NewMapFromGoMap 将 Go map 的键和值转换为动态值映射；迭代顺序不保证稳定。
func NewMapFromGoMap[K comparable, V any](m map[K]V) (Map, error) {
	ret := Map{
		Entries: make(generic.UnorderedSliceMap[Variant, Variant], 0, len(m)),
	}

	for k, v := range m {
		varK, err := ToVariant(k)
		if err != nil {
			return Map{}, err
		}

		varV, err := ToVariant(v)
		if err != nil {
			return Map{}, err
		}

		ret.Entries = append(ret.Entries, generic.UnorderedKV[Variant, Variant]{K: varK, V: varV})
	}

	return ret, nil
}

// NewMapFromSliceMap 按 SliceMap 的顺序转换键值对。
func NewMapFromSliceMap[K cmp.Ordered, V any](m generic.SliceMap[K, V]) (Map, error) {
	ret := Map{
		Entries: make(generic.UnorderedSliceMap[Variant, Variant], 0, len(m)),
	}

	for i := range m {
		kv := &m[i]

		varK, err := ToVariant(&kv.K)
		if err != nil {
			return Map{}, err
		}

		varV, err := ToVariant(&kv.V)
		if err != nil {
			return Map{}, err
		}

		ret.Entries = append(ret.Entries, generic.UnorderedKV[Variant, Variant]{K: varK, V: varV})
	}

	return ret, nil
}

// NewMapFromUnorderedSliceMap 按当前存储顺序转换键值对。
func NewMapFromUnorderedSliceMap[K comparable, V any](m generic.UnorderedSliceMap[K, V]) (Map, error) {
	ret := Map{
		Entries: make(generic.UnorderedSliceMap[Variant, Variant], 0, len(m)),
	}

	for i := range m {
		kv := &m[i]

		varK, err := ToVariant(&kv.K)
		if err != nil {
			return Map{}, err
		}

		varV, err := ToVariant(&kv.V)
		if err != nil {
			return Map{}, err
		}

		ret.Entries = append(ret.Entries, generic.UnorderedKV[Variant, Variant]{K: varK, V: varV})
	}

	return ret, nil
}

// Map 以无序切片映射保存动态键值对。
type Map struct {
	// Entries 保存键值对，允许使用可比较的 Variant 作为键。
	Entries generic.UnorderedSliceMap[Variant, Variant]
}

// Read 将动态值映射编码到 p。
func (v Map) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if err := bs.WriteUvarint(uint64(v.Entries.Len())); err != nil {
		return bs.BytesWritten(), err
	}

	for i := range v.Entries {
		kv := &v.Entries[i]
		if _, err := binaryutil.CopyToByteStream(&bs, kv.K); err != nil {
			return bs.BytesWritten(), err
		}
		if _, err := binaryutil.CopyToByteStream(&bs, kv.V); err != nil {
			return bs.BytesWritten(), err
		}
	}

	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码动态值映射。
func (v *Map) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	l, err := bs.ReadUvarint()
	if err != nil {
		return bs.BytesRead(), err
	}

	v.Entries = make([]generic.UnorderedKV[Variant, Variant], 0, min(l, 256))

	for i := uint64(0); i < l; i++ {
		var kv generic.UnorderedKV[Variant, Variant]
		if _, err := bs.WriteTo(&kv.K); err != nil {
			return bs.BytesRead(), err
		}
		if _, err := bs.WriteTo(&kv.V); err != nil {
			return bs.BytesRead(), err
		}
		v.Entries = append(v.Entries, kv)
	}

	return bs.BytesRead(), nil
}

// Size 返回映射编码后的字节数。
func (v Map) Size() int {
	n := binaryutil.SizeofUvarint(uint64(v.Entries.Len()))
	for i := range v.Entries {
		kv := &v.Entries[i]
		n += kv.K.Size()
		n += kv.V.Size()
	}
	return n
}

// TypeId 返回映射的内置类型 ID。
func (Map) TypeId() TypeId {
	return TypeId_Map
}

// Indirect 返回映射值本身。
func (v Map) Indirect() any {
	return v
}
