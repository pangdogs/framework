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

func SizeofInt8() int {
	return 1
}

func SizeofInt16() int {
	return 2
}

func SizeofInt32() int {
	return 4
}

func SizeofInt64() int {
	return 8
}

func SizeofUint8() int {
	return 1
}

func SizeofUint16() int {
	return 2
}

func SizeofUint32() int {
	return 4
}

func SizeofUint64() int {
	return 8
}

func SizeofFloat() int {
	return SizeofUint32()
}

func SizeofDouble() int {
	return SizeofUint64()
}

func SizeofByte() int {
	return 1
}

func SizeofBool() int {
	return 1
}

func SizeofBytes(v []byte) int {
	l := uint64(len(v))
	return SizeofUvarint(l) + len(v)
}

func SizeofString(v string) int {
	l := uint64(len(v))
	return SizeofUvarint(l) + len(v)
}

func SizeofBytes16() int {
	return 16
}

func SizeofBytes32() int {
	return 32
}

func SizeofBytes64() int {
	return 64
}

func SizeofBytes128() int {
	return 128
}

func SizeofBytes160() int {
	return 160
}

func SizeofBytes256() int {
	return 256
}

func SizeofBytes512() int {
	return 512
}

func SizeofVarint(v int64) int {
	uv := uint64(v) << 1
	if v < 0 {
		uv = ^uv
	}
	return SizeofUvarint(uv)
}

func SizeofUvarint(v uint64) int {
	i := 0
	for v >= 0x80 {
		v >>= 7
		i++
	}
	return i + 1
}
