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

package binaryutil

// SizeofInt8 至 SizeofBytes512 是对应定长编码的字节数。
const (
	SizeofInt8     = 1
	SizeofInt16    = 2
	SizeofInt32    = 4
	SizeofInt64    = 8
	SizeofUint8    = 1
	SizeofUint16   = 2
	SizeofUint32   = 4
	SizeofUint64   = 8
	SizeofFloat    = 4
	SizeofDouble   = 8
	SizeofByte     = 1
	SizeofBool     = 1
	SizeofBytes16  = 16
	SizeofBytes32  = 32
	SizeofBytes64  = 64
	SizeofBytes128 = 128
	SizeofBytes160 = 160
	SizeofBytes256 = 256
	SizeofBytes512 = 512
)

// SizeofBytes 返回 WriteBytes 编码 v 所需的字节数，包括长度前缀。
func SizeofBytes(v []byte) int {
	l := uint64(len(v))
	return SizeofUvarint(l) + len(v)
}

// SizeofString 返回 WriteString 编码 v 所需的字节数，包括长度前缀。
func SizeofString(v string) int {
	l := uint64(len(v))
	return SizeofUvarint(l) + len(v)
}

// SizeofVarint 返回 v 的 binary.Varint 编码字节数。
func SizeofVarint(v int64) int {
	uv := uint64(v) << 1
	if v < 0 {
		uv = ^uv
	}
	return SizeofUvarint(uv)
}

// SizeofUvarint 返回 v 的 binary.Uvarint 编码字节数。
func SizeofUvarint(v uint64) int {
	i := 0
	for v >= 0x80 {
		v >>= 7
		i++
	}
	return i + 1
}
