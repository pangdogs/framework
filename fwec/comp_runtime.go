package fwec

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/framework/plugins/dent"
)

// Runtime 运行时
type Runtime struct {
	Ctx runtime.Context
}

// GetDistEntities 获取分布式实体支持插件
func (rt Runtime) GetDistEntities() dent.IDistEntities {
	return dent.Using(rt.Ctx)
}

// CreateEntity 创建实体
func (rt Runtime) CreateEntity() core.EntityCreator {
	return core.CreateEntity(rt.Ctx)
}
