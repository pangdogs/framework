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

import "io"

// Null 是不携带负载的 GAP 空值。
type Null struct{}

// Read 不写入数据并立即结束。
func (Null) Read(p []byte) (int, error) {
	return 0, io.EOF
}

// Write 不读取数据。
func (Null) Write(p []byte) (int, error) {
	return 0, nil
}

// Size 始终返回零。
func (Null) Size() int {
	return 0
}

// TypeID 返回空值的内置类型 ID。
func (Null) TypeID() TypeID {
	return TypeID_Null
}

// Indirect 返回 nil。
func (Null) Indirect() any {
	return nil
}
