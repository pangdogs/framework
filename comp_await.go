package framework

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
)

// Await 异步等待结果返回
func (c *ComponentBehavior) Await(asyncRet ...runtime.AsyncRet) AwaitDirector {
	return AwaitDirector{
		iec:      c,
		director: core.Await(c.GetRuntime().Ctx, asyncRet...),
	}
}
