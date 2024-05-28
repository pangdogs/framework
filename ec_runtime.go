package framework

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/framework/plugins/dentr"
	"git.golaxy.org/framework/plugins/rpcstack"
)

// Runtime 运行时
type Runtime struct {
	Ctx runtime.Context
}

// GetDistEntityRegistry 获取分布式实体注册支持插件
func (rt Runtime) GetDistEntityRegistry() dentr.IDistEntityRegistry {
	return dentr.Using(rt.Ctx)
}

// GetRPCStack 获取RPC调用堆栈支持插件
func (rt Runtime) GetRPCStack() rpcstack.IRPCStack {
	return rpcstack.Using(rt.Ctx)
}

// CreateEntity 创建实体
func (rt Runtime) CreateEntity(prototype string) core.EntityCreator {
	return core.CreateEntity(rt.Ctx, prototype)
}
