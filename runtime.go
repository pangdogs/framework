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

// GetRuntime 返回 provider 所属的 framework 运行时实例。
func GetRuntime(provider runtime.CurrentContextProvider) IRuntime {
	return reinterpret.Cast[IRuntime](runtime.Current(provider))
}

// IRuntime 扩展 core runtime.Context，并聚合 framework 的运行时级 add-in 与构建入口。
// 除明确标注可并发的 API 外，应在所属运行时 goroutine 中访问运行时状态。
type IRuntime interface {
	runtime.Context
	// DistEntityRegistry 返回分布式实体注册 add-in；未安装时会 panic。
	DistEntityRegistry() dent.IDistEntityRegistry
	// RPCStack 返回 RPC 调用栈 add-in；未安装时会 panic。
	RPCStack() rpcstack.IRPCStack
	// Service 返回承载当前运行时的服务实例。
	Service() IService
	// MainEntity 返回与运行时生命周期绑定的主实体；主实体停用后运行时会终止。
	MainEntity() ec.Entity
	// AutoInjection 报告实体或组件激活时是否自动注入组件依赖。
	AutoInjection() bool
	// BuildEntity 创建绑定当前运行时及 prototype 名称的 core 实体构建器。
	BuildEntity(prototype string) *core.EntityCreator
	// L 返回当前运行时的结构化日志器。
	L() *zap.Logger
	// S 返回当前运行时的 SugaredLogger。
	S() *zap.SugaredLogger
}

type iRuntime interface {
	setMainEntity(entity ec.Entity)
	setAutoInjection(b bool)
}

// RuntimeBehavior 提供 IRuntime 的默认实现，供自定义运行时匿名嵌入。
type RuntimeBehavior struct {
	runtime.ContextBehavior
	mainEntity    ec.Entity
	autoInjection bool
}

// DistEntityRegistry 返回分布式实体注册 add-in；未安装时会 panic。
func (rt *RuntimeBehavior) DistEntityRegistry() dent.IDistEntityRegistry {
	return addins.Dentr.Require(rt)
}

// RPCStack 返回 RPC 调用栈 add-in；未安装时会 panic。
func (rt *RuntimeBehavior) RPCStack() rpcstack.IRPCStack {
	return addins.RPCStack.Require(rt)
}

// Service 返回承载当前运行时的服务实例。
func (rt *RuntimeBehavior) Service() IService {
	return reinterpret.Cast[IService](service.Current(rt))
}

// MainEntity 返回与运行时生命周期绑定的主实体；主实体停用后运行时会终止。
func (rt *RuntimeBehavior) MainEntity() ec.Entity {
	return rt.mainEntity
}

// AutoInjection 报告实体或组件激活时是否自动注入组件依赖。
func (rt *RuntimeBehavior) AutoInjection() bool {
	return rt.autoInjection
}

// BuildEntity 创建绑定当前运行时及 prototype 名称的 core 实体构建器。
func (rt *RuntimeBehavior) BuildEntity(prototype string) *core.EntityCreator {
	return core.BuildEntity(runtime.Current(rt), prototype)
}

// L 返回当前运行时的结构化日志器。
func (rt *RuntimeBehavior) L() *zap.Logger {
	return log.L(rt)
}

// S 返回当前运行时的 SugaredLogger。
func (rt *RuntimeBehavior) S() *zap.SugaredLogger {
	return log.S(rt)
}

func (rt *RuntimeBehavior) setMainEntity(entity ec.Entity) {
	rt.mainEntity = entity
}

func (rt *RuntimeBehavior) setAutoInjection(b bool) {
	rt.autoInjection = b
}
