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

const (
	TypeId_None TypeId = iota
	TypeId_Int
	TypeId_Int8
	TypeId_Int16
	TypeId_Int32
	TypeId_Int64
	TypeId_Uint
	TypeId_Uint8
	TypeId_Uint16
	TypeId_Uint32
	TypeId_Uint64
	TypeId_Float
	TypeId_Double
	TypeId_Byte
	TypeId_Bool
	TypeId_Bytes
	TypeId_String
	TypeId_Null
	TypeId_Array
	TypeId_Map
	TypeId_Error
	TypeId_CallChain
	TypeId_Customize = 32 // 自定义类型起点
)
