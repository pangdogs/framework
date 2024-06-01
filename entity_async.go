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
func (e *EntityBehavior) Async(fun generic.FuncVar0[any, async.Ret], va ...any) async.AsyncRet {
	return core.Async(e, func(_ runtime.Context, a ...any) async.Ret {
		if !e.IsAlive() {
			return async.MakeRet(nil, ErrEntityNotAlive)
		}
		return fun.Exec(a...)
	}, va...)
}

// AsyncVoid 异步执行代码，无返回值
func (e *EntityBehavior) AsyncVoid(fun generic.ActionVar0[any], va ...any) async.AsyncRet {
	return core.AsyncVoid(e, func(_ runtime.Context, a ...any) {
		if !e.IsAlive() {
			return
		}
		fun.Exec(a...)
	}, va...)
}

// Go 使用新线程执行代码，有返回值（注意线程安全）
func (e *EntityBehavior) Go(fun generic.FuncVar0[any, async.Ret], va ...any) async.AsyncRet {
	return core.Go(e.GetRuntime(), func(_ context.Context, a ...any) async.Ret {
		return fun.Exec(a...)
	}, va...)
}

// GoVoid 使用新线程执行代码，无返回值（注意线程安全）
func (e *EntityBehavior) GoVoid(fun generic.ActionVar0[any], va ...any) async.AsyncRet {
	return core.GoVoid(e.GetRuntime(), func(_ context.Context, a ...any) {
		fun.Exec(a...)
	}, va...)
}

// TimeAfter 定时器，指定时长
func (e *EntityBehavior) TimeAfter(dur time.Duration) async.AsyncRet {
	return core.TimeAfter(e.GetRuntime(), dur)
}

// TimeAt 定时器，指定时间点
func (e *EntityBehavior) TimeAt(at time.Time) async.AsyncRet {
	return core.TimeAt(e.GetRuntime(), at)
}

// TimeTick 心跳器
func (e *EntityBehavior) TimeTick(dur time.Duration) async.AsyncRet {
	return core.TimeTick(e.GetRuntime(), dur)
}
