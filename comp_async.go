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
	ErrComponentNotAlive = errors.New("async/await: component not alive")
)

// Async 异步执行代码，有返回值
func (c *ComponentBehavior) Async(fun generic.FuncVar0[any, async.Ret], va ...any) async.AsyncRet {
	return core.Async(c, func(_ runtime.Context, a ...any) async.Ret {
		if !c.IsAlive() {
			return async.MakeRet(nil, ErrComponentNotAlive)
		}
		return fun.Exec(a...)
	}, va...)
}

// AsyncVoid 异步执行代码，无返回值
func (c *ComponentBehavior) AsyncVoid(fun generic.ActionVar0[any], va ...any) async.AsyncRet {
	return core.AsyncVoid(c, func(_ runtime.Context, a ...any) {
		if !c.IsAlive() {
			return
		}
		fun.Exec(a...)
	}, va...)
}

// Go 使用新线程执行代码，有返回值（注意线程安全）
func (c *ComponentBehavior) Go(fun generic.FuncVar0[any, async.Ret], va ...any) async.AsyncRet {
	return core.Go(c.GetRuntime(), func(_ context.Context, a ...any) async.Ret {
		return fun.Exec(a...)
	}, va...)
}

// GoVoid 使用新线程执行代码，无返回值（注意线程安全）
func (c *ComponentBehavior) GoVoid(fun generic.ActionVar0[any], va ...any) async.AsyncRet {
	return core.GoVoid(c.GetRuntime(), func(_ context.Context, a ...any) {
		fun.Exec(a...)
	}, va...)
}

// TimeAfter 定时器，指定时长
func (c *ComponentBehavior) TimeAfter(dur time.Duration) async.AsyncRet {
	return core.TimeAfter(c.GetRuntime(), dur)
}

// TimeAt 定时器，指定时间点
func (c *ComponentBehavior) TimeAt(at time.Time) async.AsyncRet {
	return core.TimeAt(c.GetRuntime(), at)
}

// TimeTick 心跳器
func (c *ComponentBehavior) TimeTick(dur time.Duration) async.AsyncRet {
	return core.TimeTick(c.GetRuntime(), dur)
}
