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

// Null builtin null
type Null struct{}

// Read implements io.Reader
func (Null) Read(p []byte) (int, error) {
	return 0, nil
}

// Write implements io.Writer
func (Null) Write(p []byte) (int, error) {
	return 0, nil
}

// Size 大小
func (Null) Size() int {
	return 0
}

// TypeId 类型
func (Null) TypeId() TypeId {
	return TypeId_Null
}

// Indirect 原始值
func (Null) Indirect() any {
	return nil
}
