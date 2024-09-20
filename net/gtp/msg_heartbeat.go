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

// Heartbeat消息标志位
const (
	Flag_Ping Flag = 1 << (iota + Flag_Customize) // 心跳ping
	Flag_Pong                                     // 心跳pong
)

// MsgHeartbeat 心跳，消息体为空，可以不解析
type MsgHeartbeat struct{}

// Read implements io.Reader
func (MsgHeartbeat) Read(p []byte) (int, error) {
	return 0, io.EOF
}

// Write implements io.Writer
func (MsgHeartbeat) Write(p []byte) (int, error) {
	return 0, nil
}

// Size 大小
func (MsgHeartbeat) Size() int {
	return 0
}

// MsgId 消息Id
func (MsgHeartbeat) MsgId() MsgId {
	return MsgId_Heartbeat
}

// Clone 克隆消息对象
func (m MsgHeartbeat) Clone() MsgReader {
	return m
}
