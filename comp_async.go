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

// Submit 将 fun 提交到组件所属 Runtime，并返回执行结果 Future。
// 执行时组件若已失活，Future 返回 ErrAsyncCallerNotAlive。
func (c *ComponentBehavior) Submit(fun generic.FuncVar1[IRuntime, any, async.Result], args ...any) async.Future {
	return core.Submit(c, func(ctx runtime.Context, args ...any) async.Result {
		if !c.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		return fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), args...)
	}, args...)
}

// SubmitVoid 将无业务返回值的 fun 提交到组件所属 Runtime。
// 执行时组件若已失活，Future 返回 ErrAsyncCallerNotAlive。
func (c *ComponentBehavior) SubmitVoid(fun generic.ActionVar1[IRuntime, any], args ...any) async.Future {
	return core.Submit(c, func(ctx runtime.Context, args ...any) async.Result {
		if !c.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), args...)
		return async.NewResult(nil, nil)
	}, args...)
}

// Post 将 fun 投递到组件所属 Runtime，不创建 Future。
// fun 执行前组件已经失活时会被忽略。
func (c *ComponentBehavior) Post(fun generic.ActionVar1[IRuntime, any], args ...any) error {
	return core.Post(c, func(ctx runtime.Context, args ...any) {
		if !c.isAlive() {
			return
		}
		fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), args...)
	}, args...)
}

// Spawn 在组件生命周期 Scope 中启动后台任务。
// fun 不得直接并发访问 Runtime 局部状态。
func (c *ComponentBehavior) Spawn(fun generic.FuncVar1[context.Context, any, async.Result], args ...any) async.Future {
	return core.Spawn(c, func(ctx context.Context, args ...any) async.Result {
		return fun.UnsafeCall(ctx, args...)
	}, args...)
}

// SpawnVoid 在组件生命周期 Scope 中启动无业务返回值的后台任务。
func (c *ComponentBehavior) SpawnVoid(fun generic.ActionVar1[context.Context, any], args ...any) async.Future {
	return core.SpawnVoid(c, func(ctx context.Context, args ...any) {
		fun.UnsafeCall(ctx, args...)
	}, args...)
}

// After 在组件生命周期内等待 dur 后以当前时间完成 Future。
func (c *ComponentBehavior) After(dur time.Duration) async.Future {
	return core.After(c.AsyncScope().Context(), dur)
}

// At 在组件生命周期内等待到 at 后以当前时间完成 Future。
func (c *ComponentBehavior) At(at time.Time) async.Future {
	return core.At(c.AsyncScope().Context(), at)
}

// Every 在组件生命周期内按 dur 周期产出当前时间。
func (c *ComponentBehavior) Every(dur time.Duration) async.Stream {
	return core.Every(c.AsyncScope().Context(), dur)
}

// FromChan 将 ch 转换为绑定组件生命周期的 Stream。
func (c *ComponentBehavior) FromChan(ch <-chan any) async.Stream {
	return core.FromChan(c.AsyncScope().Context(), ch)
}

// ContinueOn 在 future 完成后回到组件所属 Runtime 执行 fun。
func (c *ComponentBehavior) ContinueOn(future async.Future, fun generic.FuncVar2[IRuntime, async.Result, any, async.Result], args ...any) async.Future {
	return core.ContinueOn(c, future, func(ctx runtime.Context, ret async.Result, args ...any) async.Result {
		if !c.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		return fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), ret, args...)
	}, args...)
}

// ContinueOnVoid 在 future 完成后回到组件所属 Runtime 执行无业务返回值的 fun。
func (c *ComponentBehavior) ContinueOnVoid(future async.Future, fun generic.ActionVar2[IRuntime, async.Result, any], args ...any) async.Future {
	return core.ContinueOn(c, future, func(ctx runtime.Context, ret async.Result, args ...any) async.Result {
		if !c.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), ret, args...)
		return async.NewResult(nil, nil)
	}, args...)
}

func (c *ComponentBehavior) isAlive() bool {
	return c.State() <= ec.ComponentState_Alive
}
