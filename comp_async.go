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
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"time"
)

// CallAsync 异步执行代码，有返回值
func (c *ComponentBehavior) CallAsync(fun generic.FuncVar0[any, async.Ret], args ...any) async.AsyncRet {
	return core.CallAsync(c, func(_ runtime.Context, args ...any) async.Ret {
		if !c.GetLiving() {
			return async.MakeRet(nil, ErrEntityOrComponentNotLiving)
		}
		return fun.UnsafeCall(args...)
	}, args...)
}

// CallVoidAsync 异步执行代码，无返回值
func (c *ComponentBehavior) CallVoidAsync(fun generic.ActionVar0[any], args ...any) async.AsyncRet {
	return core.CallAsync(c, func(_ runtime.Context, args ...any) async.Ret {
		if !c.GetLiving() {
			return async.MakeRet(nil, ErrEntityOrComponentNotLiving)
		}
		fun.UnsafeCall(args...)
		return async.VoidRet
	}, args...)
}

// GoAsync 使用新线程执行代码，有返回值（注意线程安全）
func (c *ComponentBehavior) GoAsync(fun generic.FuncVar0[any, async.Ret], args ...any) async.AsyncRet {
	return core.GoAsync(c, func(ctx context.Context, args ...any) async.Ret {
		return fun.UnsafeCall(args...)
	}, args...)
}

// GoVoidAsync 使用新线程执行代码，无返回值（注意线程安全）
func (c *ComponentBehavior) GoVoidAsync(fun generic.ActionVar0[any], args ...any) async.AsyncRet {
	return core.GoVoidAsync(c, func(ctx context.Context, args ...any) {
		fun.UnsafeCall(args...)
	}, args...)
}

// TimeAfterAsync 定时器，指定时长
func (c *ComponentBehavior) TimeAfterAsync(dur time.Duration) async.AsyncRet {
	return core.TimeAfterAsync(c, dur)
}

// TimeAtAsync 定时器，指定时间点
func (c *ComponentBehavior) TimeAtAsync(at time.Time) async.AsyncRet {
	return core.TimeAtAsync(c, at)
}

// TimeTickAsync 心跳器
func (c *ComponentBehavior) TimeTickAsync(dur time.Duration) async.AsyncRet {
	return core.TimeTickAsync(c, dur)
}

// ReadChanAsync 读取channel
func (c *ComponentBehavior) ReadChanAsync(ch <-chan any) async.AsyncRet {
	return core.ReadChanAsync(c, ch)
}
