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

package gtp

import "io"

const (
	// Flag_Ping 表示心跳探测。
	Flag_Ping Flag = 1 << (iota + Flag_Customize)
	// Flag_Pong 表示心跳响应。
	Flag_Pong
)

// MsgHeartbeat 是无消息体的心跳探测或响应。
type MsgHeartbeat struct{}

// Read 不写入数据并立即结束。
func (MsgHeartbeat) Read(p []byte) (int, error) {
	return 0, io.EOF
}

// Write 不读取数据。
func (*MsgHeartbeat) Write(p []byte) (int, error) {
	return 0, nil
}

// Size 始终返回零。
func (MsgHeartbeat) Size() int {
	return 0
}

// MsgId 返回心跳消息的内置类型 ID。
func (MsgHeartbeat) MsgId() MsgId {
	return MsgId_Heartbeat
}

// Clone 返回消息副本。
func (m MsgHeartbeat) Clone() Msg {
	return &m
}
