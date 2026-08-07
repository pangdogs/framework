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

import "io"

// CopyToBuff 调用 reader.Read 一次，将数据读入 p，并将 io.EOF 视为成功。
// 它不会循环填满 p。
func CopyToBuff[T io.Reader](p []byte, reader T) (int64, error) {
	n, err := reader.Read(p)
	if err == io.EOF {
		err = nil
	}
	return int64(n), err
}

// CopyToByteStream 调用 reader.Read 一次，将数据读入 bs 的未写区域并推进写游标。
// 它不会循环填满剩余区域，并将 io.EOF 视为成功。
func CopyToByteStream[T io.Reader](bs *ByteStream, reader T) (int64, error) {
	n, err := reader.Read(bs.wp)
	if err == io.EOF {
		err = nil
	}
	bs.wp = bs.wp[n:]
	return int64(n), err
}
