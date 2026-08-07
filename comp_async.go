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

// CallAsync 将 fun 提交到组件所属运行时执行，并返回承载调用结果的 Future。
// 执行时组件若已失活，Future 返回 ErrAsyncCallerNotAlive。
func (c *ComponentBehavior) CallAsync(fun generic.FuncVar1[IRuntime, any, async.Result], args ...any) async.Future {
	return core.CallAsync(c, func(ctx runtime.Context, args ...any) async.Result {
		if !c.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		return fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), args...)
	}, args...)
}

// CallVoidAsync 将无返回值的 fun 提交到组件所属运行时执行，并返回完成信号。
// 执行时组件若已失活，Future 返回 ErrAsyncCallerNotAlive。
func (c *ComponentBehavior) CallVoidAsync(fun generic.ActionVar1[IRuntime, any], args ...any) async.Future {
	return core.CallAsync(c, func(ctx runtime.Context, args ...any) async.Result {
		if !c.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), args...)
		return async.NewResult(nil, nil)
	}, args...)
}

// GoAsync 在新的 goroutine 中执行 fun，并返回承载调用结果的 Future。
// 传入的上下文跟随组件所属实体的生命周期；fun 不得直接并发访问运行时状态。
func (c *ComponentBehavior) GoAsync(fun generic.FuncVar1[context.Context, any, async.Result], args ...any) async.Future {
	return core.GoAsync(c.Entity(), func(ctx context.Context, args ...any) async.Result {
		return fun.UnsafeCall(ctx, args...)
	}, args...)
}

// GoVoidAsync 在新的 goroutine 中执行无返回值的 fun，并返回完成信号。
// 传入的上下文跟随组件所属实体的生命周期；fun 不得直接并发访问运行时状态。
func (c *ComponentBehavior) GoVoidAsync(fun generic.ActionVar1[context.Context, any], args ...any) async.Future {
	return core.GoVoidAsync(c.Entity(), func(ctx context.Context, args ...any) {
		fun.UnsafeCall(ctx, args...)
	}, args...)
}

// TimeAfterAsync 在 dur 后产出一次当前时间；实体失活时直接结束。
func (c *ComponentBehavior) TimeAfterAsync(dur time.Duration) async.Future {
	return core.TimeAfterAsync(c.Entity(), dur)
}

// TimeAtAsync 在 at 到达时产出一次当前时间；实体失活时直接结束。
func (c *ComponentBehavior) TimeAtAsync(at time.Time) async.Future {
	return core.TimeAtAsync(c.Entity(), at)
}

// TimeTickAsync 按 dur 周期持续产出当前时间，直到实体失活。
func (c *ComponentBehavior) TimeTickAsync(dur time.Duration) async.Future {
	return core.TimeTickAsync(c.Entity(), dur)
}

// ReadChanAsync 将 ch 中的值转换为连续产出，直到通道关闭或实体失活。
func (c *ComponentBehavior) ReadChanAsync(ch <-chan any) async.Future {
	return core.ReadChanAsync(c.Entity(), ch)
}

// Await 创建与组件关联的等待分发器；nil Future 会被忽略。
func (c *ComponentBehavior) Await(futures ...async.Future) AwaitDirector {
	return AwaitDirector{
		caller:   c,
		director: core.Await(c, futures...),
	}
}

func (c *ComponentBehavior) isAlive() bool {
	return c.State() <= ec.ComponentState_Alive
}
