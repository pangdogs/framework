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

// Package gtp defines the Golaxy Transfer Protocol.
//
// GTP 面向长连接和实时通信场景，运行在 TCP 或 WebSocket 之上，负责提供：
//   - 握手协商、链路鉴权与可选的数据加密
//   - 压缩、时钟同步、心跳和控制消息
//   - 可靠的消息时序控制、断线续连与传输状态管理
//   - 可扩展的消息类型、编解码器，以及 transport/method/sign 等配套子包
//
// 关于安全性：
//   - 当前协议支持 ECDHE、签名与验证，但不提供证书校验。
//   - 对安全要求极高的场景，建议直接在 TCP/WebSocket 下层使用 TLS，
//     并关闭本协议自带的数据加密选项。
//
// 关于性能：
//   - 消息中的 []byte 和 string 字段应尽量通过 ReadBytesRef / ReadStringRef
//     这类引用式读取方法访问，以减少额外拷贝。
package gtp
