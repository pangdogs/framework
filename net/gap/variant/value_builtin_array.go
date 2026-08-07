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
	// ErrSnapshotReadonly 表示试图向只读数组快照解码数据。
	ErrSnapshotReadonly = fmt.Errorf("%w: snapshot array is readonly", ErrVariant)
)

// NewArray 将 Go 切片逐项转换为动态值数组。
func NewArray[T any](arr []T) (Array, error) {
	ret := Array{
		Items: make([]Variant, 0, len(arr)),
	}

	for i := range arr {
		item, err := ToVariant(&arr[i])
		if err != nil {
			return Array{}, err
		}
		ret.Items = append(ret.Items, item)
	}

	return ret, nil
}

// Array 保存有序动态值；快照形态只保留已编码字节且不可写入。
type Array struct {
	Items         []Variant        // 非快照形态的数组项。
	IsSnapshot    bool             // 是否为只读编码快照。
	SnapshotBytes binaryutil.Bytes // 快照编码；可回收快照使用完后必须释放。
}

// Read 将数组或其快照编码到 p。
func (v Array) Read(p []byte) (int, error) {
	if v.IsSnapshot {
		data := v.SnapshotBytes.Payload()
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

// Write 从 p 解码数组；快照形态返回 ErrSnapshotReadonly。
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

// Size 返回数组编码后的字节数。
func (v Array) Size() int {
	if v.IsSnapshot {
		return len(v.SnapshotBytes.Payload())
	}
	n := binaryutil.SizeofUvarint(uint64(len(v.Items)))
	for i := range v.Items {
		n += v.Items[i].Size()
	}
	return n
}

// TypeId 返回数组的内置类型 ID。
func (Array) TypeId() TypeId {
	return TypeId_Array
}

// Indirect 返回数组值本身。
func (v Array) Indirect() any {
	return v
}

// Snapshot 创建数组编码的只读副本；recyclable 为 true 时使用完必须调用 ReleaseIfSnapshot。
func (v Array) Snapshot(recyclable bool) (Array, error) {
	if v.IsSnapshot {
		return Array{
			IsSnapshot:    true,
			SnapshotBytes: binaryutil.CloneBytes(recyclable, v.SnapshotBytes.Payload()),
		}, nil
	}

	data := binaryutil.NewBytes(recyclable, v.Size())
	ret := Array{
		IsSnapshot:    true,
		SnapshotBytes: data,
	}

	if _, err := binaryutil.CopyToBuff(data.Payload(), v); err != nil {
		data.Release()
		return Array{}, err
	}

	return ret, nil
}

// ReleaseIfSnapshot 释放快照持有的可回收字节缓冲区；非快照调用无效。
func (v Array) ReleaseIfSnapshot() {
	if v.IsSnapshot {
		v.SnapshotBytes.Release()
	}
}
