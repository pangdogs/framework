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

// BytesWriter 将数据顺序写入固定字节切片，且不允许一次写入越过剩余空间。
type BytesWriter struct {
	N     int    // 已写入的字节数。
	Bytes []byte // 目标字节切片。
}

// NewBytesWriter 创建从 bs 起始位置写入的固定缓冲区 writer。
func NewBytesWriter(bs []byte) *BytesWriter {
	return &BytesWriter{
		N:     0,
		Bytes: bs,
	}
}

// Write 将 p 完整写入剩余空间；空间不足时不写入任何数据并返回 ErrLimitReached。
func (l *BytesWriter) Write(p []byte) (int, error) {
	if l.N >= len(l.Bytes) {
		return 0, ErrLimitReached
	}

	// 超出限制时保持全有或全无语义。
	if len(p) > len(l.Bytes)-l.N {
		return 0, ErrLimitReached
	}

	copy(l.Bytes[l.N:], p)
	l.N += len(p)

	return len(p), nil
}
