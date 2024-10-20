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

package gap

import (
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/types"
	"hash/fnv"
	"reflect"
)

// MsgId 消息Id
type MsgId = uint32

// MakeMsgId 创建类型Id
func MakeMsgId(msg Msg) MsgId {
	hash := fnv.New32a()
	rt := reflect.ValueOf(msg).Elem().Type()
	if rt.PkgPath() == "" || rt.Name() == "" {
		exception.Panic("gap: unsupported type")
	}
	hash.Write([]byte(types.FullNameRT(rt)))
	return MsgId(MsgId_Customize + hash.Sum32())
}

// MakeMsgIdT 创建类型Id
func MakeMsgIdT[T any]() MsgId {
	hash := fnv.New32a()
	rt := reflect.TypeFor[T]()
	if rt.PkgPath() == "" || rt.Name() == "" || !reflect.PointerTo(rt).Implements(reflect.TypeFor[Msg]()) {
		exception.Panic("gap: unsupported type")
	}
	hash.Write([]byte(types.FullNameRT(rt)))
	return MsgId(MsgId_Customize + hash.Sum32())
}
