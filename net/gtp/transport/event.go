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
	// ErrEvent 是 GTP 传输事件处理错误的根错误。
	ErrEvent = errors.New("gtp-event")
	// ErrIncorrectMsg 表示事件消息与要求的具体类型不匹配。
	ErrIncorrectMsg = fmt.Errorf("%w: incorrect msg type", ErrEvent)
)

// IEvent 是消息类型擦除后的 GTP 传输事件。
type IEvent = Event[gtp.ReadableMsg]

// Event 组合消息头的传输元数据与具体消息。
type Event[T gtp.ReadableMsg] struct {
	Flags gtp.Flags // 消息标志位。
	Seq   uint32    // 消息序号。
	Ack   uint32    // 对端确认序号。
	Msg   T         // 具体消息。
}

// Interface 将具体事件转换为类型擦除事件；消息未实现 gtp.Msg 时 panic。
func (e Event[T]) Interface() IEvent {
	msg, ok := any(e.Msg).(gtp.Msg)
	if !ok {
		msg, ok = any(&e.Msg).(gtp.Msg)
		if !ok {
			exception.Panic(ErrIncorrectMsg)
			panic("unreachable")
		}
	}
	return IEvent{
		Flags: e.Flags,
		Seq:   e.Seq,
		Ack:   e.Ack,
		Msg:   msg,
	}
}

// AssertEvent 将类型擦除事件断言为具体消息事件；类型不匹配时 panic。
func AssertEvent[T gtp.ReadableMsg](e IEvent) Event[T] {
	ret := Event[T]{
		Flags: e.Flags,
		Seq:   e.Seq,
		Ack:   e.Ack,
	}

	{
		msg, ok := any(e.Msg).(T)
		if ok {
			ret.Msg = msg
			return ret
		}
	}

	{
		msg, ok := any(e.Msg).(*T)
		if ok {
			ret.Msg = *msg
			return ret
		}
	}

	exception.Panic(ErrIncorrectMsg)
	panic("unreachable")
}
