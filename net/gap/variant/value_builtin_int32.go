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

// Int32 是 GAP 的 int32 动态值。
type Int32 int32

// Read 将值编码到 p。
func (v Int32) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteInt32(int32(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码值。
func (v *Int32) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadInt32()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Int32(val)
	return bs.BytesRead(), nil
}

// Size 返回值编码后的字节数。
func (Int32) Size() int {
	return binaryutil.SizeofInt32
}

// TypeId 返回 int32 的内置类型 ID。
func (Int32) TypeId() TypeId {
	return TypeId_Int32
}

// Indirect 返回 int32。
func (v Int32) Indirect() any {
	return int32(v)
}
