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

// Byte 是 GAP 的 byte 动态值。
type Byte byte

// Read 将值编码到 p。
func (v Byte) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteByte(byte(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码值。
func (v *Byte) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadByte()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Byte(val)
	return bs.BytesRead(), nil
}

// Size 返回值编码后的字节数。
func (Byte) Size() int {
	return binaryutil.SizeofByte
}

// TypeID 返回 byte 的内置类型 ID。
func (Byte) TypeID() TypeID {
	return TypeID_Byte
}

// Indirect 返回 byte。
func (v Byte) Indirect() any {
	return byte(v)
}
