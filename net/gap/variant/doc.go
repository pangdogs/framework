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
// 包内以 Variant 作为统一入口，Variant 持有 TypeId 和对应的可读值。
// 常用内置类型会在初始化时注册，包括整数、浮点、布尔、字节串、字符串、
// Null、Array、Map、Error 和 CallChain。自定义类型需要实现 Value 接口，
// 并通过 VariantCreator().Declare 注册后，才能根据 TypeId 反序列化。
//
// 常用入口包括：
//   - NewVariant / CastVariant：把值包装为普通 Variant。
//   - NewSerializedVariant / CastSerializedVariant：把值转换为显式序列化变体。
//   - NewSerializedArray / NewSerializedMapFrom...：创建由 SerializedVariant
//     元素组成的序列化容器。
//   - NewSerializedValue：缓存单个值的编码字节。
//   - GenTypeId / GenTypeIdT：为自定义类型生成稳定的 TypeId。
//
// 当前序列化相关结构主要包括：
//   - SerializedValue：单个值的已编码字节。
//   - SerializedArray：数组结果的序列化容器。
//   - SerializedMap：映射结果的序列化容器。
//   - SerializedVariant：统一包装上述序列化值或容器。
//
// Ref 返回底层普通值视图，用于继续走现有的 Variant / Array / Map 编码链路。
// Release 用于归还序列化过程中持有的缓存资源。调用方需要自行保证同一底层
// 序列化对象只释放一次，不要对共享底层资源的多个包装对象重复调用 Release。
// 当前这套序列化结构主要用于 RPC 返回值回包链，尤其是需要跨异步等待后再写入
// MsgRPCReply 的场景。
package variant
