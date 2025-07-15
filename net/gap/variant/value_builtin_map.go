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
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
)

// MakeMapFromGoMap 创建map
func MakeMapFromGoMap[K comparable, V any](m map[K]V) (Map, error) {
	varMap := make(Map, 0, len(m))

	for k, v := range m {
		varK, err := CastVariant(k)
		if err != nil {
			return nil, err
		}

		varV, err := CastVariant(v)
		if err != nil {
			return nil, err
		}

		varMap.ToUnorderedSliceMap().Add(varK, varV)
	}

	return varMap, nil
}

// MakeMapFromSliceMap 创建map
func MakeMapFromSliceMap[K cmp.Ordered, V any](m generic.SliceMap[K, V]) (Map, error) {
	varMap := make(Map, 0, len(m))

	for i := range m {
		kv := &m[i]

		varK, err := CastVariant(&kv.K)
		if err != nil {
			return nil, err
		}

		varV, err := CastVariant(&kv.V)
		if err != nil {
			return nil, err
		}

		varMap.ToUnorderedSliceMap().Add(varK, varV)
	}

	return varMap, nil
}

// MakeMapFromUnorderedSliceMap 创建map
func MakeMapFromUnorderedSliceMap[K comparable, V any](m generic.UnorderedSliceMap[K, V]) (Map, error) {
	varMap := make(Map, 0, len(m))

	for i := range m {
		kv := &m[i]

		varK, err := CastVariant(&kv.K)
		if err != nil {
			return nil, err
		}

		varV, err := CastVariant(&kv.V)
		if err != nil {
			return nil, err
		}

		varMap.ToUnorderedSliceMap().Add(varK, varV)
	}

	return varMap, nil
}

// Map map
type Map generic.UnorderedSliceMap[Variant, Variant]

// Read implements io.Reader
func (v Map) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if err := bs.WriteUvarint(uint64(len(v))); err != nil {
		return bs.BytesWritten(), err
	}

	for i := range v {
		kv := &v[i]

		if _, err := binaryutil.CopyToByteStream(&bs, kv.K); err != nil {
			return bs.BytesWritten(), err
		}

		if _, err := binaryutil.CopyToByteStream(&bs, kv.V); err != nil {
			return bs.BytesWritten(), err
		}
	}

	return bs.BytesWritten(), io.EOF
}

// Write implements io.Writer
func (v *Map) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	l, err := bs.ReadUvarint()
	if err != nil {
		return bs.BytesRead(), err
	}

	*v = make([]generic.UnorderedKV[Variant, Variant], l)

	for i := uint64(0); i < l; i++ {
		kv := &(*v)[i]

		if _, err := bs.WriteTo(&kv.K); err != nil {
			return bs.BytesRead(), err
		}

		if _, err := bs.WriteTo(&kv.V); err != nil {
			return bs.BytesRead(), err
		}
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (v Map) Size() int {
	n := binaryutil.SizeofUvarint(uint64(len(v)))
	for i := range v {
		kv := &v[i]
		n += kv.K.Size()
		n += kv.V.Size()
	}
	return n
}

// TypeId 类型
func (Map) TypeId() TypeId {
	return TypeId_Map
}

// Indirect 原始值
func (v Map) Indirect() any {
	return v
}

// Release 释放资源
func (v Map) Release() {
	for i := range v {
		kv := &v[i]
		kv.K.Release()
		kv.V.Release()
	}
}

// ToUnorderedSliceMap 转换为UnorderedSliceMap
func (v *Map) ToUnorderedSliceMap() *generic.UnorderedSliceMap[Variant, Variant] {
	return (*generic.UnorderedSliceMap[Variant, Variant])(v)
}
