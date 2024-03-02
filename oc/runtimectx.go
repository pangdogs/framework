package oc

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/framework/plugins/dent"
)

// RuntimeCtx 运行时上下文
type RuntimeCtx struct {
	runtime.Context
}

// GetDistEntities 获取分布式实体支持插件
func (ctx RuntimeCtx) GetDistEntities() dent.IDistEntities {
	return dent.Using(ctx.Context)
}

// CreateEntity 创建实体
func (ctx RuntimeCtx) CreateEntity() core.EntityCreator {
	return core.CreateEntity(ctx.Context)
}
