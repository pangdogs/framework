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

// 内置动态值类型 ID。自定义类型必须从 TypeID_Customize 起分配。
const (
	// TypeID_None 表示未设置动态值类型。
	TypeID_None TypeID = iota
	// TypeID_Int 标识 int。
	TypeID_Int
	// TypeID_Int8 标识 int8。
	TypeID_Int8
	// TypeID_Int16 标识 int16。
	TypeID_Int16
	// TypeID_Int32 标识 int32。
	TypeID_Int32
	// TypeID_Int64 标识 int64。
	TypeID_Int64
	// TypeID_Uint 标识 uint。
	TypeID_Uint
	// TypeID_Uint8 标识 uint8。
	TypeID_Uint8
	// TypeID_Uint16 标识 uint16。
	TypeID_Uint16
	// TypeID_Uint32 标识 uint32。
	TypeID_Uint32
	// TypeID_Uint64 标识 uint64。
	TypeID_Uint64
	// TypeID_Float 标识 float32。
	TypeID_Float
	// TypeID_Double 标识 float64。
	TypeID_Double
	// TypeID_Byte 标识 byte。
	TypeID_Byte
	// TypeID_Bool 标识 bool。
	TypeID_Bool
	// TypeID_Bytes 标识 []byte。
	TypeID_Bytes
	// TypeID_String 标识 string。
	TypeID_String
	// TypeID_Null 标识空值。
	TypeID_Null
	// TypeID_Array 标识动态值数组。
	TypeID_Array
	// TypeID_Map 标识动态值映射。
	TypeID_Map
	// TypeID_Error 标识可传输错误。
	TypeID_Error
	// TypeID_CallChain 标识 RPC 调用链。
	TypeID_CallChain
	// TypeID_Customize 是自定义类型 ID 的起始偏移。
	TypeID_Customize = 32
)
