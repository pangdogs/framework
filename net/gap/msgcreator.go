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
	"fmt"
	"maps"
	"reflect"
	"runtime"
	"sync/atomic"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/types"
)

var (
	// ErrNotDeclared 表示指定消息 ID 尚未注册。
	ErrNotDeclared = fmt.Errorf("%w: msg not declared", ErrGAP)
)

// IMsgCreator 按消息 ID 注册并构造 GAP 消息。
type IMsgCreator interface {
	// Declare 注册消息指针的具体类型；重复 ID 会 panic。
	Declare(msg Msg)
	// New 创建指定消息 ID 对应的新消息指针。
	New(msgID MsgID) (Msg, error)
}

var msgCreator = NewMsgCreator()

// DefaultMsgCreator 返回已注册内置消息的进程级消息构建器。
func DefaultMsgCreator() IMsgCreator {
	return msgCreator
}

func init() {
	DefaultMsgCreator().Declare(&MsgRPCRequest{})
	DefaultMsgCreator().Declare(&MsgRPCReply{})
	DefaultMsgCreator().Declare(&MsgOnewayRPC{})
	DefaultMsgCreator().Declare(&MsgForward{})
}

// NewMsgCreator 创建空的并发安全消息构建器。
func NewMsgCreator() IMsgCreator {
	return &_MsgCreator{}
}

// _MsgCreator 使用写时复制快照保存消息类型映射。
type _MsgCreator struct {
	msgTypes atomic.Pointer[map[MsgID]reflect.Type]
}

// Declare 以消息 ID 注册消息的元素类型。
func (c *_MsgCreator) Declare(msg Msg) {
	if msg == nil {
		exception.Panicf("%w: %w: msg is nil", ErrGAP, core.ErrArgs)
	}

	for {
		var m map[MsgID]reflect.Type

		old := c.msgTypes.Load()
		if old != nil {
			m = maps.Clone(*old)
		}

		if m == nil {
			m = make(map[MsgID]reflect.Type)
		}

		if rtype, ok := (m)[msg.MsgID()]; ok {
			exception.Panicf("%w: msg(%d) has already been declared by %q; rename the message type or return a different MsgID", ErrGAP, msg.MsgID(), types.FullNameRT(rtype))
		}

		m[msg.MsgID()] = reflect.TypeOf(msg).Elem()

		if c.msgTypes.CompareAndSwap(old, &m) {
			break
		}

		runtime.Gosched()
	}
}

// New 根据当前类型快照创建新的消息指针。
func (c *_MsgCreator) New(msgID MsgID) (Msg, error) {
	m := c.msgTypes.Load()
	if m == nil || *m == nil {
		return nil, ErrNotDeclared
	}

	rtype, ok := (*m)[msgID]
	if !ok {
		return nil, ErrNotDeclared
	}

	return reflect.New(rtype).Interface().(Msg), nil
}
