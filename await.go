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
	ErrAsyncCallerNotAlive = errors.New("async/await: async caller is not alive")
)

type iAsyncCaller interface {
	isAlive() bool
}

// AwaitDirector 异步等待分发器
type AwaitDirector struct {
	caller   iAsyncCaller
	director core.AwaitDirector
}

// Any 异步等待任意一个结果返回，有返回值
func (ad AwaitDirector) Any(fun generic.FuncVar2[IRuntime, async.Result, any, async.Result], args ...any) async.Future {
	return ad.director.Any(func(ctx runtime.Context, ret async.Result, args ...any) async.Result {
		if !ad.caller.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		return fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), ret, args...)
	}, args...)
}

// AnyVoid 异步等待任意一个结果返回，无返回值
func (ad AwaitDirector) AnyVoid(fun generic.ActionVar2[IRuntime, async.Result, any], args ...any) async.Future {
	return ad.director.Any(func(ctx runtime.Context, ret async.Result, args ...any) async.Result {
		if !ad.caller.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), ret, args...)
		return async.NewResult(nil, nil)
	}, args...)
}

// OK 异步等待任意一个结果成功返回，有返回值
func (ad AwaitDirector) OK(fun generic.FuncVar2[IRuntime, async.Result, any, async.Result], args ...any) async.Future {
	return ad.director.OK(func(ctx runtime.Context, ret async.Result, args ...any) async.Result {
		if !ad.caller.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		return fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), ret, args...)
	}, args...)
}

// OKVoid 异步等待任意一个结果成功返回，无返回值
func (ad AwaitDirector) OKVoid(fun generic.ActionVar2[IRuntime, async.Result, any], args ...any) async.Future {
	return ad.director.OK(func(ctx runtime.Context, ret async.Result, args ...any) async.Result {
		if !ad.caller.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), ret, args...)
		return async.NewResult(nil, nil)
	}, args...)
}

// All 异步等待所有结果返回，有返回值
func (ad AwaitDirector) All(fun generic.FuncVar2[IRuntime, []async.Result, any, async.Result], args ...any) async.Future {
	return ad.director.All(func(ctx runtime.Context, rets []async.Result, args ...any) async.Result {
		if !ad.caller.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		return fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), rets, args...)
	}, args...)
}

// AllVoid 异步等待所有结果返回，无返回值
func (ad AwaitDirector) AllVoid(fun generic.ActionVar2[IRuntime, []async.Result, any], args ...any) async.Future {
	return ad.director.All(func(ctx runtime.Context, rets []async.Result, args ...any) async.Result {
		if !ad.caller.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), rets, args...)
		return async.NewResult(nil, nil)
	}, args...)
}

// Transform 异步等待产出（yield）返回，并变换结果
func (ad AwaitDirector) Transform(fun generic.FuncVar2[IRuntime, async.Result, any, async.Result], args ...any) async.Future {
	return ad.director.Transform(func(ctx runtime.Context, ret async.Result, args ...any) async.Result {
		if !ad.caller.isAlive() {
			return async.NewResult(nil, ErrAsyncCallerNotAlive)
		}
		return fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), ret, args...)
	}, args...)
}

// Foreach 异步等待产出（yield）返回
func (ad AwaitDirector) Foreach(fun generic.ActionVar2[IRuntime, async.Result, any], args ...any) async.Future {
	return ad.director.Foreach(func(ctx runtime.Context, ret async.Result, args ...any) {
		if !ad.caller.isAlive() {
			return
		}
		fun.UnsafeCall(reinterpret.Cast[IRuntime](ctx), ret, args...)
	}, args...)
}
