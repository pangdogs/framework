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

// String 是 GAP 的字符串动态值。
type String string

// Read 将长度前缀和字符串内容编码到 p。
func (v String) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteString(string(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码字符串。
func (v *String) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = String(val)
	return bs.BytesRead(), nil
}

// Size 返回长度前缀和字符串内容的总字节数。
func (v String) Size() int {
	return binaryutil.SizeofString(string(v))
}

// TypeID 返回字符串的内置类型 ID。
func (String) TypeID() TypeID {
	return TypeID_String
}

// Indirect 返回 string。
func (v String) Indirect() any {
	return string(v)
}
