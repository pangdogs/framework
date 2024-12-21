/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package framework

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/framework/addins/dentr"
	"git.golaxy.org/framework/addins/rpcstack"
)

// IRuntimeInstantiation 运行时实例化接口
type IRuntimeInstantiation interface {
	Instantiation() IRuntimeInstance
}

// GetRuntimeInstance 获取运行时实例
func GetRuntimeInstance(provider runtime.CurrentContextProvider) IRuntimeInstance {
	return reinterpret.Cast[IRuntimeInstance](runtime.Current(provider))
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
