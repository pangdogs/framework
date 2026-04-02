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
	"time"

	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/reinterpret"
)

// CallAsync 异步执行代码，有返回值
func (e *EntityBehavior) CallAsync(fun generic.FuncVar1[IRuntime, any, async.Result], args ...any) async.Future {
	return core.CallAsync(e, func(ctx runtime.Context, args ...any) async.Result {
		if !e.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		return fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), args...)
	}, args...)
}

// CallVoidAsync 异步执行代码，无返回值
func (e *EntityBehavior) CallVoidAsync(fun generic.ActionVar1[IRuntime, any], args ...any) async.Future {
	return core.CallAsync(e, func(ctx runtime.Context, args ...any) async.Result {
		if !e.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), args...)
		return async.NewResult(nil, nil)
	}, args...)
}

// GoAsync 使用新线程执行代码，有返回值（注意线程安全）
func (e *EntityBehavior) GoAsync(fun generic.FuncVar1[context.Context, any, async.Result], args ...any) async.Future {
	return core.GoAsync(e, func(ctx context.Context, args ...any) async.Result {
		return fun.UnsafeCall(ctx, args...)
	}, args...)
}

// GoVoidAsync 使用新线程执行代码，无返回值（注意线程安全）
func (e *EntityBehavior) GoVoidAsync(fun generic.ActionVar1[context.Context, any], args ...any) async.Future {
	return core.GoVoidAsync(e, func(ctx context.Context, args ...any) {
		fun.UnsafeCall(ctx, args...)
	}, args...)
}

// TimeAfterAsync 定时器，指定时长
func (e *EntityBehavior) TimeAfterAsync(dur time.Duration) async.Future {
	return core.TimeAfterAsync(e, dur)
}

// TimeAtAsync 定时器，指定时间点
func (e *EntityBehavior) TimeAtAsync(at time.Time) async.Future {
	return core.TimeAtAsync(e, at)
}

// TimeTickAsync 心跳器
func (e *EntityBehavior) TimeTickAsync(dur time.Duration) async.Future {
	return core.TimeTickAsync(e, dur)
}

// ReadChanAsync 读取channel
func (e *EntityBehavior) ReadChanAsync(ch <-chan any) async.Future {
	return core.ReadChanAsync(e, ch)
}

// Await 异步等待结果返回
func (e *EntityBehavior) Await(futures ...async.Future) AwaitDirector {
	return AwaitDirector{
		caller:   e,
		director: core.Await(e, futures...),
	}
}

func (e *EntityBehavior) isAlive() bool {
	return e.State() <= ec.EntityState_Alive
}
