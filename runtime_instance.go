package framework

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/framework/plugins/dentr"
	"git.golaxy.org/framework/plugins/rpcstack"
)

// IRuntimeInstantiation 运行时实例化接口
type IRuntimeInstantiation interface {
	Instantiation() IRuntimeInstance
}

// IRuntimeInstance 运行时实例接口
type IRuntimeInstance interface {
	runtime.Context
	// GetDistEntityRegistry 获取分布式实体注册支持
	GetDistEntityRegistry() dentr.IDistEntityRegistry
	// GetRPCStack 获取RPC调用堆栈支持
	GetRPCStack() rpcstack.IRPCStack
	// GetService 获取服务实例
	GetService() IServiceInstance
	// CreateEntity 创建实体
	CreateEntity(prototype string) core.EntityCreator
}

// RuntimeInstance 运行时实例
type RuntimeInstance struct {
	runtime.ContextBehavior
}

// GetDistEntityRegistry 获取分布式实体注册支持
func (inst *RuntimeInstance) GetDistEntityRegistry() dentr.IDistEntityRegistry {
	return dentr.Using(inst)
}

// GetRPCStack 获取RPC调用堆栈支持
func (inst *RuntimeInstance) GetRPCStack() rpcstack.IRPCStack {
	return rpcstack.Using(inst)
}

// GetService 获取服务
func (inst *RuntimeInstance) GetService() IServiceInstance {
	return reinterpret.Cast[IServiceInstance](service.Current(inst))
}

// CreateEntity 创建实体
func (inst *RuntimeInstance) CreateEntity(prototype string) core.EntityCreator {
	return core.CreateEntity(runtime.Current(inst), prototype)
}
