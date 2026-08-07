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

const (
	// MsgId_None 表示未设置消息类型。
	MsgId_None MsgId = iota
	// MsgId_RPC_Request 标识需要响应的 RPC 请求。
	MsgId_RPC_Request
	// MsgId_RPC_Reply 标识 RPC 响应。
	MsgId_RPC_Reply
	// MsgId_OnewayRPC 标识无需响应的 RPC 通知。
	MsgId_OnewayRPC
	// MsgId_Forward 标识封装其他消息的路由转发消息。
	MsgId_Forward
	// MsgId_Customize 是自定义消息 ID 的起始偏移。
	MsgId_Customize = 32
)
