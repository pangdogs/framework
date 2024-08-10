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

// Async 异步执行代码，有返回值
func (e *EntityBehavior) Async(fun generic.FuncVar0[any, async.Ret], args ...any) async.AsyncRet {
	return core.Async(e, func(runtime.Context, ...any) async.Ret {
		if !e.IsAlive() {
			return async.MakeRet(nil, ErrEntityNotAlive)
		}
		return fun.Exec(args...)
	})
}

// AsyncVoid 异步执行代码，无返回值
func (e *EntityBehavior) AsyncVoid(fun generic.ActionVar0[any], args ...any) async.AsyncRet {
	return core.AsyncVoid(e, func(runtime.Context, ...any) {
		if !e.IsAlive() {
			return
		}
		fun.Exec(args...)
	})
}

// Go 使用新线程执行代码，有返回值（注意线程安全）
func (e *EntityBehavior) Go(fun generic.FuncVar0[any, async.Ret], args ...any) async.AsyncRet {
	return core.Go(e, func(context.Context, ...any) async.Ret {
		return fun.Exec(args...)
	})
}

// GoVoid 使用新线程执行代码，无返回值（注意线程安全）
func (e *EntityBehavior) GoVoid(fun generic.ActionVar0[any], args ...any) async.AsyncRet {
	return core.GoVoid(e, func(context.Context, ...any) {
		fun.Exec(args...)
	})
}

// TimeAfter 定时器，指定时长
func (e *EntityBehavior) TimeAfter(dur time.Duration) async.AsyncRet {
	return core.TimeAfter(e, dur)
}

// TimeAt 定时器，指定时间点
func (e *EntityBehavior) TimeAt(at time.Time) async.AsyncRet {
	return core.TimeAt(e, at)
}

// TimeTick 心跳器
func (e *EntityBehavior) TimeTick(dur time.Duration) async.AsyncRet {
	return core.TimeTick(e, dur)
}
