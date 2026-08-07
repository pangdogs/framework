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
	"errors"

	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/reinterpret"
)

var (
	// ErrAsyncCallerNotAlive 表示等待结果返回时，发起等待的实体或组件已经失活。
	ErrAsyncCallerNotAlive = errors.New("async/await: async caller is not alive")
)

type iAsyncCaller interface {
	isAlive() bool
}

// AwaitDirector 组合一个或多个 Future，并把后续回调调度回调用方所属运行时。
// 回调执行前会检查发起等待的实体或组件是否仍然存活。
type AwaitDirector struct {
	caller   iAsyncCaller
	director core.AwaitDirector
}

// Any 等待首个 Future 返回结果，再在所属运行时执行 fun。
// 结果是否包含错误不影响选择。
func (ad AwaitDirector) Any(fun generic.FuncVar2[IRuntime, async.Result, any, async.Result], args ...any) async.Future {
	return ad.director.Any(func(ctx runtime.Context, ret async.Result, args ...any) async.Result {
		if !ad.caller.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		return fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), ret, args...)
	}, args...)
}

// AnyVoid 等待首个 Future 返回结果，再在所属运行时执行无返回值的 fun。
func (ad AwaitDirector) AnyVoid(fun generic.ActionVar2[IRuntime, async.Result, any], args ...any) async.Future {
	return ad.director.Any(func(ctx runtime.Context, ret async.Result, args ...any) async.Result {
		if !ad.caller.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), ret, args...)
		return async.NewResult(nil, nil)
	}, args...)
}

// OK 等待首个成功结果，再在所属运行时执行 fun。
// 所有 Future 均失败时返回 core.ErrNoFutureSucceeded。
func (ad AwaitDirector) OK(fun generic.FuncVar2[IRuntime, async.Result, any, async.Result], args ...any) async.Future {
	return ad.director.OK(func(ctx runtime.Context, ret async.Result, args ...any) async.Result {
		if !ad.caller.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		return fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), ret, args...)
	}, args...)
}

// OKVoid 等待首个成功结果，再在所属运行时执行无返回值的 fun。
// 所有 Future 均失败时返回 core.ErrNoFutureSucceeded。
func (ad AwaitDirector) OKVoid(fun generic.ActionVar2[IRuntime, async.Result, any], args ...any) async.Future {
	return ad.director.OK(func(ctx runtime.Context, ret async.Result, args ...any) async.Result {
		if !ad.caller.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), ret, args...)
		return async.NewResult(nil, nil)
	}, args...)
}

// All 按传入顺序收集所有 Future 的结果，再在所属运行时执行 fun。
func (ad AwaitDirector) All(fun generic.FuncVar2[IRuntime, []async.Result, any, async.Result], args ...any) async.Future {
	return ad.director.All(func(ctx runtime.Context, rets []async.Result, args ...any) async.Result {
		if !ad.caller.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		return fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), rets, args...)
	}, args...)
}

// AllVoid 按传入顺序收集所有 Future 的结果，再在所属运行时执行无返回值的 fun。
func (ad AwaitDirector) AllVoid(fun generic.ActionVar2[IRuntime, []async.Result, any], args ...any) async.Future {
	return ad.director.All(func(ctx runtime.Context, rets []async.Result, args ...any) async.Result {
		if !ad.caller.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), rets, args...)
		return async.NewResult(nil, nil)
	}, args...)
}

// Transform 合并所有 Future 的连续产出，在所属运行时逐项执行 fun，并产出转换结果。
func (ad AwaitDirector) Transform(fun generic.FuncVar2[IRuntime, async.Result, any, async.Result], args ...any) async.Future {
	return ad.director.Transform(func(ctx runtime.Context, ret async.Result, args ...any) async.Result {
		if !ad.caller.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		return fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), ret, args...)
	}, args...)
}

// Foreach 合并所有 Future 的连续产出，并在所属运行时逐项执行 fun。
func (ad AwaitDirector) Foreach(fun generic.ActionVar2[IRuntime, async.Result, any], args ...any) async.Future {
	return ad.director.Foreach(func(ctx runtime.Context, ret async.Result, args ...any) {
		if !ad.caller.isAlive() {
			return
		}
		fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), ret, args...)
	}, args...)
}
