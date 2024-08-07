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

// Package gtp Golaxy传输层协议（Golaxy Transfer Protocol），适用于长连接、实时通信的工作场景，需要工作在可靠网络协议（TCP/WebSocket）之上，支持链路加密、链路鉴权、断线续连等特性。
/*
	- 关于加密，支持秘钥交换（ECDHE）、签名与验证，不支持证书验证，对于安全性要求极高的应用场景，应该使用TLS协议直接加密链路，并关闭本协议的数据加密选项。
	- 关于断线续联，支持下层协议断连后，双端缓存待发送的消息包，待使用新连接重连后继续收发消息。
	- 支持消息包压缩。
	- 支持可靠的消息包传输时序。
	- 支持新增自定义消息。
	- 协议适用于网络游戏、远程控制和传感器交互等实时性要求较高的场景。
	- 为了提高性能，本层所有消息中[]byte和string类型字段，应该使用ReadBytesRef和ReadStringRef读取。
*/
package gtp
