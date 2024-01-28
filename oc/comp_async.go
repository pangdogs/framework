package oc

import (
	"context"
	"errors"
	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/util/generic"
	"time"
)

var (
	ErrComponentIsDead = errors.New("async/await: component is dead")
)

// Async 异步执行代码，有返回值
func (c *ComponentBehavior) Async(fun generic.FuncVar0[any, runtime.Ret], va ...any) runtime.AsyncRet {
	return core.Async(c.GetRtCtx(), func(_ runtime.Context, a ...any) runtime.Ret {
		if c.GetState() > ec.ComponentState_Living {
			return runtime.MakeRet(nil, ErrComponentIsDead)
		}
		return fun.Exec(a...)
	}, va...)
}

// AsyncVoid 异步执行代码，无返回值
func (c *ComponentBehavior) AsyncVoid(fun generic.ActionVar0[any], va ...any) runtime.AsyncRet {
	return core.AsyncVoid(c.GetRtCtx(), func(_ runtime.Context, a ...any) {
		if c.GetState() > ec.ComponentState_Living {
			return
		}
		fun.Exec(a...)
	}, va...)
}

// Go 使用新线程执行代码，有返回值（注意线程安全）
func (c *ComponentBehavior) Go(fun generic.FuncVar0[any, runtime.Ret], va ...any) runtime.AsyncRet {
	return core.Go(c.GetRtCtx(), func(_ context.Context, a ...any) runtime.Ret {
		return fun.Exec(a...)
	}, va...)
}

// GoVoid 使用新线程执行代码，无返回值（注意线程安全）
func (c *ComponentBehavior) GoVoid(fun generic.ActionVar0[any], va ...any) runtime.AsyncRet {
	return core.GoVoid(c.GetRtCtx(), func(_ context.Context, a ...any) {
		fun.Exec(a...)
	}, va...)
}

// TimeAfter 定时器，指定时长
func (c *ComponentBehavior) TimeAfter(dur time.Duration) runtime.AsyncRet {
	return core.TimeAfter(c.GetRtCtx(), dur)
}

// TimeAt 定时器，指定时间点
func (c *ComponentBehavior) TimeAt(at time.Time) runtime.AsyncRet {
	return core.TimeAt(c.GetRtCtx(), at)
}

// TimeTick 心跳器
func (c *ComponentBehavior) TimeTick(dur time.Duration) runtime.AsyncRet {
	return core.TimeTick(c.GetRtCtx(), dur)
}
