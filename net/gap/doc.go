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

// Package gap defines the Golaxy Application Protocol.
//
// GAP 运行在 GTP 或消息队列之上，负责承载应用层消息，适合服务到服务、
// 服务到客户端、以及路由转发等通信场景。当前包提供：
//   - 统一的消息接口、消息头和消息创建器
//   - Forward、RPC request/reply、oneway RPC 等基础消息模型
//   - 序列化与反序列化入口
//   - 配套的 codec 与 variant 子包，用于编解码和动态类型参数传输
//
// 当需要在稳定传输层之上表达业务消息、RPC 参数或可扩展载荷时，应优先使用 GAP。
package gap
