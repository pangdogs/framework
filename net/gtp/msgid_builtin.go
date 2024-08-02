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
	MsgId_None                   MsgId = iota // 未设置
	MsgId_Hello                               // Hello Handshake C<->S 不加密
	MsgId_ECDHESecretKeyExchange              // ECDHE秘钥交换 Handshake S<->C 不加密
	MsgId_ChangeCipherSpec                    // 变更密码规范 Handshake S<->C 不加密
	MsgId_Auth                                // 鉴权 Handshake C->S 加密
	MsgId_Continue                            // 重连 Handshake C->S 加密
	MsgId_Finished                            // 握手结束 Handshake S<->C 加密
	MsgId_Rst                                 // 重置链路 Ctrl S->C 加密
	MsgId_Heartbeat                           // 心跳 Ctrl C<->S or S<->C 加密
	MsgId_SyncTime                            // 时钟同步 Ctrl C<->S 加密
	MsgId_Payload                             // 数据传输 Trans C<->S or S<->C 加密
	MsgId_Customize              = 16         // 自定义消息起点
)
