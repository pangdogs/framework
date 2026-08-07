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
	"io"

	"git.golaxy.org/framework/utils/binaryutil"
)

// Bytes 是 GAP 的字节切片动态值。
type Bytes []byte

// Read 将长度前缀和字节内容编码到 p。
func (v Bytes) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteBytes(v); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码并复制字节内容。
func (v *Bytes) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadBytes()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = val
	return bs.BytesRead(), nil
}

// Size 返回长度前缀和内容的总字节数。
func (v Bytes) Size() int {
	return binaryutil.SizeofBytes(v)
}

// TypeId 返回字节切片的内置类型 ID。
func (Bytes) TypeId() TypeId {
	return TypeId_Bytes
}

// Indirect 返回底层字节切片，不执行复制。
func (v Bytes) Indirect() any {
	return []byte(v)
}
