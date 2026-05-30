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

// Package variant 提供 GAP 消息和 RPC 负载使用的动态值模型。
//
// Variant 是统一的协议值包装，持有 TypeId 和对应的可读值。内置值包括整数、
// 浮点数、布尔值、字节串、字符串、Null、Array、Map、Error 和 CallChain。
// 自定义值需要实现 Value 接口，并通过 VariantCreator 注册后，才能根据 TypeId
// 反序列化。
//
// 常用入口：
//   - NewVariant：把已有 ReadableValue 包装为 Variant。
//   - ToVariant：把常见 Go 值转换为 Variant。
//   - Variant.ToNative：把反序列化后的 GAP 值转换为指定 Go reflect.Type。
//   - Array.Snapshot：冻结 Array 的编码载荷，用于延迟交付或跨 goroutine
//     交付后的写出。快照 Array 是只读形态，只用于后续编码。
package variant
