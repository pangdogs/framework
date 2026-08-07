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

// SerializedMsg 将已有字节切片作为 GAP 消息体发送，不复制 Data。
type SerializedMsg struct {
	Id   MsgId  // 消息类型 ID。
	Data []byte // 已编码消息体；由调用方维护其生命周期。
}

// Read 将已编码消息体复制到 p。
func (m SerializedMsg) Read(p []byte) (int, error) {
	if len(p) < len(m.Data) {
		return 0, io.ErrShortWrite
	}
	return copy(p, m.Data), io.EOF
}

// Size 返回已编码消息体的字节数。
func (m SerializedMsg) Size() int {
	return len(m.Data)
}

// MsgId 返回消息类型 ID。
func (m SerializedMsg) MsgId() MsgId {
	return m.Id
}
