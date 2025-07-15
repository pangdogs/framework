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

package gap

import (
	"io"
)

// SerializedMsg 已序列化消息
type SerializedMsg struct {
	Id   MsgId  // 消息Id
	Data []byte // 消息内容（引用）
}

// Read implements io.Reader
func (m SerializedMsg) Read(p []byte) (int, error) {
	if len(p) < len(m.Data) {
		return 0, io.ErrShortWrite
	}
	return copy(p, m.Data), io.EOF
}

// Size 大小
func (m SerializedMsg) Size() int {
	return len(m.Data)
}

// MsgId 消息Id
func (m SerializedMsg) MsgId() MsgId {
	return m.Id
}
