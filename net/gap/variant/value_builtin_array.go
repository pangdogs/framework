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
	"fmt"
	"io"

	"git.golaxy.org/framework/utils/binaryutil"
)

var (
	ErrSnapshotReadonly = fmt.Errorf("%w: snapshot array is readonly", ErrVariant)
)

// NewArray 创建array
func NewArray[T any](arr []T) (Array, error) {
	ret := Array{
		Items: make([]Variant, 0, len(arr)),
	}

	for i := range arr {
		item, err := CastVariant(&arr[i])
		if err != nil {
			return Array{}, err
		}
		ret.Items = append(ret.Items, item)
	}

	return ret, nil
}

// Array array
type Array struct {
	Items      []Variant
	IsSnapshot bool
	Snapshot   binaryutil.Bytes
}

// Read implements io.Reader
func (v Array) Read(p []byte) (int, error) {
	if v.IsSnapshot {
		data := v.Snapshot.Payload()
		if len(p) < len(data) {
			return 0, io.ErrShortWrite
		}
		return copy(p, data), io.EOF
	}

	bs := binaryutil.NewBigEndianStream(p)

	if err := bs.WriteUvarint(uint64(len(v.Items))); err != nil {
		return bs.BytesWritten(), err
	}

	for i := range v.Items {
		if _, err := binaryutil.CopyToByteStream(&bs, v.Items[i]); err != nil {
			return bs.BytesWritten(), err
		}
	}

	return bs.BytesWritten(), io.EOF
}

// Write implements io.Writer
func (v *Array) Write(p []byte) (int, error) {
	if v.IsSnapshot {
		return 0, ErrSnapshotReadonly
	}

	bs := binaryutil.NewBigEndianStream(p)

	l, err := bs.ReadUvarint()
	if err != nil {
		return bs.BytesRead(), err
	}

	v.Items = make([]Variant, 0, min(l, 256))

	for i := uint64(0); i < l; i++ {
		var item Variant
		if _, err := bs.WriteTo(&item); err != nil {
			return bs.BytesRead(), err
		}
		v.Items = append(v.Items, item)
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (v Array) Size() int {
	if v.IsSnapshot {
		return len(v.Snapshot.Payload())
	}
	n := binaryutil.SizeofUvarint(uint64(len(v.Items)))
	for i := range v.Items {
		n += v.Items[i].Size()
	}
	return n
}

// TypeId 类型
func (Array) TypeId() TypeId {
	return TypeId_Array
}

// Indirect 原始值
func (v Array) Indirect() any {
	return v
}

// ToSnapshot 转换为快照
func (v Array) ToSnapshot(recyclable bool) (Array, error) {
	if v.IsSnapshot {
		return Array{
			IsSnapshot: true,
			Snapshot:   binaryutil.CloneBytes(recyclable, v.Snapshot.Payload()),
		}, nil
	}

	data := binaryutil.NewBytes(recyclable, v.Size())
	ret := Array{
		IsSnapshot: true,
		Snapshot:   data,
	}

	if _, err := binaryutil.CopyToBuff(data.Payload(), v); err != nil {
		data.Release()
		return Array{}, err
	}

	return ret, nil
}

// ReleaseIfSnapshot 释放快照字节对象
func (v Array) ReleaseIfSnapshot() {
	if v.IsSnapshot {
		v.Snapshot.Release()
	}
}
