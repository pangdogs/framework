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

// Package variant provides the dynamic value system used by GAP messages and RPC
// payloads.
//
// 这个包把可传输值统一抽象为带 TypeId 的 Variant，主要用于：
//   - GAP RPC 参数与返回值
//   - 自定义消息中的动态字段
//   - Map、Array、Error、CallChain 等通用复合值的跨网络传输
//
// 包内会在初始化时注册常用内置类型，包括整数、浮点、布尔、字节串、字符串、
// Null、Map、Array、Error 和 CallChain。自定义类型需要实现 Value 接口，
// 并通过 VariantCreator().Declare 注册后，才能根据 TypeId 反序列化。
//
// 常见入口包括：
//   - NewVariant / CastVariant：把值包装为 Variant
//   - NewSerializedVariant / CastSerializedVariant：把值提前序列化后再包装
//   - NewSerializedValue：生成仅保留 TypeId 和原始字节的序列化值
//   - GenTypeId / GenTypeIdT：为自定义类型生成稳定的 TypeId
//
// 如果 Variant 持有的是 SerializedValue 或其他带缓冲区的对象，使用方在值不再
// 需要时应调用 Release，以便及时归还底层资源。
package variant
