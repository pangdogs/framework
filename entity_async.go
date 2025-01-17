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

package framework

import (
	"context"
	"errors"
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"time"
)

var (
	ErrEntityNotAlive = errors.New("async/await: entity not alive")
)

// CallAsync 异步执行代码，有返回值
func (e *EntityBehavior) CallAsync(fun generic.FuncVar0[any, async.Ret], args ...any) async.AsyncRet {
	return core.CallAsync(e, func(runtime.Context, ...any) async.Ret {
		if !e.IsAlive() {
			return async.MakeRet(nil, ErrEntityNotAlive)
		}
		return fun.UnsafeCall(args...)
	})
}

// CallVoidAsync 异步执行代码，无返回值
func (e *EntityBehavior) CallVoidAsync(fun generic.ActionVar0[any], args ...any) async.AsyncRet {
	return core.CallVoidAsync(e, func(runtime.Context, ...any) {
		if !e.IsAlive() {
			return
		}
		fun.UnsafeCall(args...)
	})
}

// GoAsync 使用新线程执行代码，有返回值（注意线程安全）
func (e *EntityBehavior) GoAsync(fun generic.FuncVar0[any, async.Ret], args ...any) async.AsyncRet {
	return core.GoAsync(e, func(context.Context, ...any) async.Ret {
		return fun.UnsafeCall(args...)
	})
}

// GoVoidAsync 使用新线程执行代码，无返回值（注意线程安全）
func (e *EntityBehavior) GoVoidAsync(fun generic.ActionVar0[any], args ...any) async.AsyncRet {
	return core.GoVoidAsync(e, func(context.Context, ...any) {
		fun.UnsafeCall(args...)
	})
}

// TimeAfterAsync 定时器，指定时长
func (e *EntityBehavior) TimeAfterAsync(dur time.Duration) async.AsyncRet {
	return core.TimeAfterAsync(e, dur)
}

// TimeAtAsync 定时器，指定时间点
func (e *EntityBehavior) TimeAtAsync(at time.Time) async.AsyncRet {
	return core.TimeAtAsync(e, at)
}

// TimeTickAsync 心跳器
func (e *EntityBehavior) TimeTickAsync(dur time.Duration) async.AsyncRet {
	return core.TimeTickAsync(e, dur)
}
