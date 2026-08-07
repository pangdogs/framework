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

package variant

// 内置动态值类型 ID。自定义类型必须从 TypeId_Customize 起分配。
const (
	// TypeId_None 表示未设置动态值类型。
	TypeId_None TypeId = iota
	// TypeId_Int 标识 int。
	TypeId_Int
	// TypeId_Int8 标识 int8。
	TypeId_Int8
	// TypeId_Int16 标识 int16。
	TypeId_Int16
	// TypeId_Int32 标识 int32。
	TypeId_Int32
	// TypeId_Int64 标识 int64。
	TypeId_Int64
	// TypeId_Uint 标识 uint。
	TypeId_Uint
	// TypeId_Uint8 标识 uint8。
	TypeId_Uint8
	// TypeId_Uint16 标识 uint16。
	TypeId_Uint16
	// TypeId_Uint32 标识 uint32。
	TypeId_Uint32
	// TypeId_Uint64 标识 uint64。
	TypeId_Uint64
	// TypeId_Float 标识 float32。
	TypeId_Float
	// TypeId_Double 标识 float64。
	TypeId_Double
	// TypeId_Byte 标识 byte。
	TypeId_Byte
	// TypeId_Bool 标识 bool。
	TypeId_Bool
	// TypeId_Bytes 标识 []byte。
	TypeId_Bytes
	// TypeId_String 标识 string。
	TypeId_String
	// TypeId_Null 标识空值。
	TypeId_Null
	// TypeId_Array 标识动态值数组。
	TypeId_Array
	// TypeId_Map 标识动态值映射。
	TypeId_Map
	// TypeId_Error 标识可传输错误。
	TypeId_Error
	// TypeId_CallChain 标识 RPC 调用链。
	TypeId_CallChain
	// TypeId_Customize 是自定义类型 ID 的起始偏移。
	TypeId_Customize = 32
)
