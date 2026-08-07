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

import (
	"errors"
	"io"
)

var (
	// ErrLimitReached 表示写入已达到或将超过配置的字节上限。
	ErrLimitReached = errors.New("i/o limit reached")
)

// LimitWriter 限制写入底层 writer 的累计字节数，并保证超限的单次写入不会触达底层 writer。
type LimitWriter struct {
	Limit int       // 允许写入的总字节数。
	N     int       // 底层 writer 已接受的累计字节数。
	W     io.Writer // 接收数据的底层 writer。
}

// NewLimitWriter 创建累计写入上限为 n 的 writer；负数 n 按零处理。
func NewLimitWriter(w io.Writer, n int) *LimitWriter {
	// 负数限制等同于禁止写入。
	if n < 0 {
		n = 0
	}
	return &LimitWriter{
		Limit: n,
		N:     0,
		W:     w,
	}
}

// Write 将 p 交给底层 writer；若 p 会使累计字节数超过 Limit，则不写入并返回 ErrLimitReached。
// 底层 writer 的短写和错误会原样返回，N 按其实际返回的字节数增加。
func (l *LimitWriter) Write(p []byte) (int, error) {
	if l.N >= l.Limit {
		return 0, ErrLimitReached
	}

	// 超出限制时不向底层 writer 写入任何数据。
	if len(p) > l.Limit-l.N {
		return 0, ErrLimitReached
	}

	n, err := l.W.Write(p)
	l.N += n
	return n, err
}
