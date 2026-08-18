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

const (
	// MsgID_None 表示未设置消息类型。
	MsgID_None MsgID = iota
	// MsgID_Hello 标识双向、明文的 Hello 握手消息。
	MsgID_Hello
	// MsgID_ECDHESecretKeyExchange 标识双向、明文的 ECDHE 密钥交换消息。
	MsgID_ECDHESecretKeyExchange
	// MsgID_ChangeCipherSpec 标识双向、明文的密码规范切换消息。
	MsgID_ChangeCipherSpec
	// MsgID_Auth 标识客户端发往服务端的加密鉴权消息。
	MsgID_Auth
	// MsgID_Continue 标识客户端发往服务端的加密会话续接消息。
	MsgID_Continue
	// MsgID_Finished 标识双向、加密的握手完成消息。
	MsgID_Finished
	// MsgID_Rst 标识服务端发往客户端的加密链路重置消息。
	MsgID_Rst
	// MsgID_Heartbeat 标识双向、加密的心跳消息。
	MsgID_Heartbeat
	// MsgID_SyncTime 标识双向、加密的时钟同步消息。
	MsgID_SyncTime
	// MsgID_Payload 标识双向、加密的业务负载消息。
	MsgID_Payload
	// MsgID_Customize 是自定义消息 ID 的起始偏移。
	MsgID_Customize = 16
)
