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
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
)

// Double builtin double
type Double float64

// Read implements io.Reader
func (v Double) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteDouble(float64(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write implements io.Writer
func (v *Double) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadDouble()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Double(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (Double) Size() int {
	return binaryutil.SizeofDouble()
}

// TypeId 类型
func (Double) TypeId() TypeId {
	return TypeId_Double
}

// Indirect 原始值
func (v Double) Indirect() any {
	return float64(v)
}
