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

package transport

import (
	"errors"
	"fmt"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/framework/net/gtp"
)

var (
	ErrEvent        = errors.New("gtp-event")                        // 消息事件错误
	ErrIncorrectMsg = fmt.Errorf("%w: incorrect msg type", ErrEvent) // 错误的消息类型
)

// IEvent 消息事件接口
type IEvent = Event[gtp.Msg]

// Event 消息事件
type Event[T gtp.Msg] struct {
	Flags gtp.Flags // 标志位
	Seq   uint32    // 消息序号
	Ack   uint32    // 应答序号
	Msg   T         // 消息
}

// Interface 接口化事件，转换为事件接口
func (e Event[T]) Interface() IEvent {
	return IEvent{
		Flags: e.Flags,
		Seq:   e.Seq,
		Ack:   e.Ack,
		Msg:   e.Msg,
	}
}

// AssertEvent 断言事件，转换为事件具体类型
func AssertEvent[T gtp.Msg](e IEvent) Event[T] {
	msg, ok := any(e.Msg).(T)
	if !ok {
		exception.Panic(ErrIncorrectMsg)
		panic("unreachable")
	}
	return Event[T]{
		Flags: e.Flags,
		Seq:   e.Seq,
		Ack:   e.Ack,
		Msg:   msg,
	}
}
