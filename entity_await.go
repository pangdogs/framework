package framework

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
)

// Await 异步等待结果返回
func (e *EntityBehavior) Await(asyncRet ...runtime.AsyncRet) AwaitDirector {
	return AwaitDirector{
		iec:      e,
		director: core.Await(e.GetRuntime().Ctx, asyncRet...),
	}
}
