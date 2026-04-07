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

// Package binaryutil 提供二进制读写和字节缓冲辅助工具。
//
// 这个包围绕 []byte 提供了常用的二进制序列化基础能力，包括：
//   - 可顺序读写的 ByteStream
//   - 可复用的字节缓冲 Bytes 与字节池
//   - 各类基础类型和定长字节块的大小计算
//   - 面向 io.Reader/io.Writer 的拷贝与限长写入辅助
//
// 它主要作为协议编解码和底层高性能字节处理的基础设施使用。
package binaryutil
