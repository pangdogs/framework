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
)

// ReadableValue 表示可编码到 GAP 动态值字节流的值。
type ReadableValue interface {
	io.Reader
	// Size 返回值编码后的字节数。
	Size() int
	// TypeId 返回动态值类型 ID。
	TypeId() TypeId
	// Indirect 返回对应的 Go 值。
	Indirect() any
}

// Value 表示既可编码也可从字节流解码的动态值。
type Value interface {
	ReadableValue
	io.Writer
}
