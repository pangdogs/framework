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
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
)

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

		if _, err := binaryutil.ReadTo(&bs, kv.K); err != nil {
			return bs.BytesWritten(), err
		}

		if _, err := binaryutil.ReadTo(&bs, kv.V); err != nil {
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

		if kv.K.Serialized() {
			kv.K.SerializedValue.Release()
		}

		if kv.V.Serialized() {
			kv.V.SerializedValue.Release()
		}
	}
}

// ToUnorderedSliceMap 转换为UnorderedSliceMap
func (v *Map) ToUnorderedSliceMap() *generic.UnorderedSliceMap[Variant, Variant] {
	return (*generic.UnorderedSliceMap[Variant, Variant])(v)
}
