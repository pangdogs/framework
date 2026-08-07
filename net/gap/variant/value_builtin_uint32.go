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

// Uint32 是 GAP 的 uint32 动态值。
type Uint32 uint32

// Read 将值编码到 p。
func (v Uint32) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint32(uint32(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码值。
func (v *Uint32) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Uint32(val)
	return bs.BytesRead(), nil
}

// Size 返回值编码后的字节数。
func (Uint32) Size() int {
	return binaryutil.SizeofUint32
}

// TypeId 返回 uint32 的内置类型 ID。
func (Uint32) TypeId() TypeId {
	return TypeId_Uint32
}

// Indirect 返回 uint32。
func (v Uint32) Indirect() any {
	return uint32(v)
}
