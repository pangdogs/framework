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

// BytesWriter will only write bytes to the underlying writer until the limit is reached.
type BytesWriter struct {
	N     int
	Bytes []byte
}

// NewBytesWriter creates a new instance of BytesWriter.
func NewBytesWriter(bs []byte) *BytesWriter {
	return &BytesWriter{
		N:     0,
		Bytes: bs,
	}
}

// Write implements io.Writer
func (l *BytesWriter) Write(p []byte) (int, error) {
	if l.N >= len(l.Bytes) {
		return 0, ErrLimitReached
	}

	// Write 0 bytes if the limit is to be exceeded.
	if len(p) > len(l.Bytes)-l.N {
		return 0, ErrLimitReached
	}

	copy(l.Bytes[l.N:], p)
	l.N += len(p)

	return len(p), nil
}
