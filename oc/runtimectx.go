package oc

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/framework/plugins/dent"
)

type RuntimeCtx struct {
	runtime.Context
}

func (ctx RuntimeCtx) GetDistEntities() dent.IDistEntities {
	return dent.Using(ctx.Context)
}
