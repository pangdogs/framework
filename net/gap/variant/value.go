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

// Value 值
type Value interface {
	ValueReader
	ValueWriter
}

// ValueReader 读取值
type ValueReader interface {
	io.Reader
	// Size 大小
	Size() int
	// TypeId 类型
	TypeId() TypeId
	// Indirect 原始值
	Indirect() any
}

// ValueWriter 写入值
type ValueWriter interface {
	io.Writer
	// Size 大小
	Size() int
	// TypeId 类型
	TypeId() TypeId
}
