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
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/framework/addins"
	"git.golaxy.org/framework/addins/dent"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/rpcstack"
	"go.uber.org/zap"
)

// GetRuntime 获取运行时实例
func GetRuntime(provider runtime.CurrentContextProvider) IRuntime {
	return reinterpret.Cast[IRuntime](runtime.Current(provider))
}

// IRuntime 运行时实例接口
type IRuntime interface {
	runtime.Context
	// DistEntityRegistry 获取分布式实体注册支持
	DistEntityRegistry() dent.IDistEntityRegistry
	// RPCStack 获取RPC调用堆栈支持
	RPCStack() rpcstack.IRPCStack
	// Service 获取服务实例
	Service() IService
	// MainEntity 获取主实体（主实体和运行时生命周期绑定，主实体销毁时，运行时将会停止运行）
	MainEntity() ec.Entity
	// AutoInjection 获取是否自动注入组件
	AutoInjection() bool
	// BuildEntity 创建实体
	BuildEntity(prototype string) *core.EntityCreator
	// L 结构化日志
	L() *zap.Logger
	// S 传统日志
	S() *zap.SugaredLogger
}

type iRuntime interface {
	setMainEntity(entity ec.Entity)
	setAutoInjection(b bool)
}

// RuntimeBehavior 运行时实例行为
type RuntimeBehavior struct {
	runtime.ContextBehavior
	mainEntity    ec.Entity
	autoInjection bool
}

// DistEntityRegistry 获取分布式实体注册支持
func (rt *RuntimeBehavior) DistEntityRegistry() dent.IDistEntityRegistry {
	return addins.Dentr.Require(rt)
}

// RPCStack 获取RPC调用堆栈支持
func (rt *RuntimeBehavior) RPCStack() rpcstack.IRPCStack {
	return addins.RPCStack.Require(rt)
}

// Service 获取服务
func (rt *RuntimeBehavior) Service() IService {
	return reinterpret.Cast[IService](service.Current(rt))
}

// MainEntity 获取主实体（主实体和运行时生命周期绑定，主实体销毁时，运行时将会停止运行）
func (rt *RuntimeBehavior) MainEntity() ec.Entity {
	return rt.mainEntity
}

// AutoInjection 获取是否自动注入组件
func (rt *RuntimeBehavior) AutoInjection() bool {
	return rt.autoInjection
}

// BuildEntity 创建实体
func (rt *RuntimeBehavior) BuildEntity(prototype string) *core.EntityCreator {
	return core.BuildEntity(runtime.Current(rt), prototype)
}

// L 结构化日志
func (rt *RuntimeBehavior) L() *zap.Logger {
	return log.L(rt)
}

// S 传统日志
func (rt *RuntimeBehavior) S() *zap.SugaredLogger {
	return log.S(rt)
}

func (rt *RuntimeBehavior) setMainEntity(entity ec.Entity) {
	rt.mainEntity = entity
}

func (rt *RuntimeBehavior) setAutoInjection(b bool) {
	rt.autoInjection = b
}
