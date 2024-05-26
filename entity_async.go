package framework

import (
	"context"
	"errors"
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/util/generic"
	"time"
)

var (
	ErrEntityNotAlive = errors.New("async/await: entity not alive")
)

// Async 异步执行代码，有返回值
func (e *EntityBehavior) Async(fun generic.FuncVar0[any, runtime.Ret], va ...any) runtime.AsyncRet {
	return core.Async(e, func(_ runtime.Context, a ...any) runtime.Ret {
		if !e.IsAlive() {
			return runtime.MakeRet(nil, ErrEntityNotAlive)
		}
		return fun.Exec(a...)
	}, va...)
}

// AsyncVoid 异步执行代码，无返回值
func (e *EntityBehavior) AsyncVoid(fun generic.ActionVar0[any], va ...any) runtime.AsyncRet {
	return core.AsyncVoid(e, func(_ runtime.Context, a ...any) {
		if !e.IsAlive() {
			return
		}
		fun.Exec(a...)
	}, va...)
}

// Go 使用新线程执行代码，有返回值（注意线程安全）
func (e *EntityBehavior) Go(fun generic.FuncVar0[any, runtime.Ret], va ...any) runtime.AsyncRet {
	return core.Go(e.GetRuntime().Ctx, func(_ context.Context, a ...any) runtime.Ret {
		return fun.Exec(a...)
	}, va...)
}

// GoVoid 使用新线程执行代码，无返回值（注意线程安全）
func (e *EntityBehavior) GoVoid(fun generic.ActionVar0[any], va ...any) runtime.AsyncRet {
	return core.GoVoid(e.GetRuntime().Ctx, func(_ context.Context, a ...any) {
		fun.Exec(a...)
	}, va...)
}

// TimeAfter 定时器，指定时长
func (e *EntityBehavior) TimeAfter(dur time.Duration) runtime.AsyncRet {
	return core.TimeAfter(e.GetRuntime().Ctx, dur)
}

// TimeAt 定时器，指定时间点
func (e *EntityBehavior) TimeAt(at time.Time) runtime.AsyncRet {
	return core.TimeAt(e.GetRuntime().Ctx, at)
}

// TimeTick 心跳器
func (e *EntityBehavior) TimeTick(dur time.Duration) runtime.AsyncRet {
	return core.TimeTick(e.GetRuntime().Ctx, dur)
}
