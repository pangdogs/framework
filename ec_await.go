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
)

var (
	ErrEntityOrComponentNotLiving = errors.New("async/await: entity or component not living")
)

// AwaitDirector 异步等待分发器
type AwaitDirector struct {
	iec      iEC
	director core.AwaitDirector
}

// Any 异步等待任意一个结果返回，有返回值
func (ad AwaitDirector) Any(fun generic.FuncVar1[async.Ret, any, async.Ret], args ...any) async.AsyncRet {
	return ad.director.Any(func(_ runtime.Context, ret async.Ret, args ...any) async.Ret {
		if !ad.iec.GetLiving() {
			return async.MakeRet(nil, ErrEntityOrComponentNotLiving)
		}
		return fun.UnsafeCall(ret, args...)
	}, args...)
}

// AnyVoid 异步等待任意一个结果返回，无返回值
func (ad AwaitDirector) AnyVoid(fun generic.ActionVar1[async.Ret, any], args ...any) async.AsyncRet {
	return ad.director.Any(func(_ runtime.Context, ret async.Ret, args ...any) async.Ret {
		if !ad.iec.GetLiving() {
			return async.MakeRet(nil, ErrEntityOrComponentNotLiving)
		}
		fun.UnsafeCall(ret, args...)
		return async.VoidRet
	}, args...)
}

// OK 异步等待任意一个结果成功返回，有返回值
func (ad AwaitDirector) OK(fun generic.FuncVar1[async.Ret, any, async.Ret], args ...any) async.AsyncRet {
	return ad.director.OK(func(_ runtime.Context, ret async.Ret, args ...any) async.Ret {
		if !ad.iec.GetLiving() {
			return async.MakeRet(nil, ErrEntityOrComponentNotLiving)
		}
		return fun.UnsafeCall(ret, args...)
	}, args...)
}

// OKVoid 异步等待任意一个结果成功返回，无返回值
func (ad AwaitDirector) OKVoid(fun generic.ActionVar1[async.Ret, any], args ...any) async.AsyncRet {
	return ad.director.OK(func(_ runtime.Context, ret async.Ret, args ...any) async.Ret {
		if !ad.iec.GetLiving() {
			return async.MakeRet(nil, ErrEntityOrComponentNotLiving)
		}
		fun.UnsafeCall(ret, args...)
		return async.VoidRet
	}, args...)
}

// All 异步等待所有结果返回，有返回值
func (ad AwaitDirector) All(fun generic.FuncVar1[[]async.Ret, any, async.Ret], args ...any) async.AsyncRet {
	return ad.director.All(func(_ runtime.Context, rets []async.Ret, args ...any) async.Ret {
		if !ad.iec.GetLiving() {
			return async.MakeRet(nil, ErrEntityOrComponentNotLiving)
		}
		return fun.UnsafeCall(rets, args...)
	}, args...)
}

// AllVoid 异步等待所有结果返回，无返回值
func (ad AwaitDirector) AllVoid(fun generic.ActionVar1[[]async.Ret, any], args ...any) async.AsyncRet {
	return ad.director.All(func(_ runtime.Context, rets []async.Ret, args ...any) async.Ret {
		if !ad.iec.GetLiving() {
			return async.MakeRet(nil, ErrEntityOrComponentNotLiving)
		}
		fun.UnsafeCall(rets, args...)
		return async.VoidRet
	}, args...)
}

// Transform 异步等待产出（yield）返回，并变换结果
func (ad AwaitDirector) Transform(fun generic.FuncVar1[async.Ret, any, async.Ret], args ...any) async.AsyncRet {
	return ad.director.Transform(func(_ runtime.Context, ret async.Ret, args ...any) async.Ret {
		if !ad.iec.GetLiving() {
			return async.MakeRet(nil, ErrEntityOrComponentNotLiving)
		}
		return fun.UnsafeCall(ret, args...)
	}, args...)
}

// Foreach 异步等待产出（yield）返回
func (ad AwaitDirector) Foreach(fun generic.ActionVar1[async.Ret, any], args ...any) async.AsyncRet {
	return ad.director.Foreach(func(_ runtime.Context, ret async.Ret, args ...any) {
		if !ad.iec.GetLiving() {
			return
		}
		fun.UnsafeCall(ret, args...)
	}, args...)
}
