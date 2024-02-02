package oc

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/util/generic"
)

// AwaitDirector 异步等待分发器
type AwaitDirector struct {
	cb       *ComponentBehavior
	director core.AwaitDirector
}

// Any 异步等待任意一个结果返回
func (ad AwaitDirector) Any(fun generic.ActionVar1[runtime.Ret, any], va ...any) {
	ad.director.Any(runtime.Current(ad.cb), func(_ runtime.Context, ret runtime.Ret, a ...any) {
		if ad.cb.GetState() > ec.ComponentState_Living {
			return
		}
		fun.Exec(ret, a...)
	}, va...)
}

// AnyOK 异步等待任意一个结果成功返回
func (ad AwaitDirector) AnyOK(fun generic.ActionVar1[runtime.Ret, any], va ...any) {
	ad.director.AnyOK(runtime.Current(ad.cb), func(_ runtime.Context, ret runtime.Ret, a ...any) {
		if ad.cb.GetState() > ec.ComponentState_Living {
			return
		}
		fun.Exec(ret, a...)
	}, va...)
}

// All 异步等待所有结果返回
func (ad AwaitDirector) All(fun generic.ActionVar1[[]runtime.Ret, any], va ...any) {
	ad.director.All(runtime.Current(ad.cb), func(_ runtime.Context, rets []runtime.Ret, a ...any) {
		if ad.cb.GetState() > ec.ComponentState_Living {
			return
		}
		fun.Exec(rets, a...)
	}, va...)
}

// Pipe 异步等待管道返回
func (ad AwaitDirector) Pipe(fun generic.ActionVar1[runtime.Ret, any], va ...any) {
	ad.director.Pipe(runtime.Current(ad.cb), func(_ runtime.Context, ret runtime.Ret, a ...any) {
		if ad.cb.GetState() > ec.ComponentState_Living {
			return
		}
		fun.Exec(ret, a...)
	}, va...)
}

// Await 异步等待结果返回
func (c *ComponentBehavior) Await(asyncRet ...runtime.AsyncRet) AwaitDirector {
	return AwaitDirector{
		cb:       c,
		director: core.Await(c.GetRuntimeCtx(), asyncRet...),
	}
}
