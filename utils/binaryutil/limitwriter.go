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
	ErrLimitReached = errors.New("i/o limit reached")
)

// LimitWriter will only write bytes to the underlying writer until the limit is reached.
type LimitWriter struct {
	Limit int
	N     int
	W     io.Writer
}

// NewLimitWriter creates a new instance of LimitWriter.
func NewLimitWriter(w io.Writer, n int) *LimitWriter {
	// If anyone tries this, just make a 0 writer.
	if n < 0 {
		n = 0
	}
	return &LimitWriter{
		Limit: n,
		N:     0,
		W:     w,
	}
}

// Write implements io.Writer
func (l *LimitWriter) Write(p []byte) (int, error) {
	if l.N >= l.Limit {
		return 0, ErrLimitReached
	}

	// Write 0 bytes if the limit is to be exceeded.
	if len(p) > l.Limit-l.N {
		return 0, ErrLimitReached
	}

	n, err := l.W.Write(p)
	l.N += n
	return n, err
}
