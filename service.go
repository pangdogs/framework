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
	"sync"
	"sync/atomic"

	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/framework/addins"
	"git.golaxy.org/framework/addins/broker"
	"git.golaxy.org/framework/addins/dent"
	"git.golaxy.org/framework/addins/discovery"
	"git.golaxy.org/framework/addins/dsvc"
	"git.golaxy.org/framework/addins/dsync"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/rpc"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// GetService 返回 provider 所属的 framework 服务实例。
func GetService(provider runtime.ConcurrentContextProvider) IService {
	return reinterpret.Cast[IService](service.Current(provider))
}

// IService 扩展 core service.Context，并聚合 framework 的服务级 add-in 与构建入口。
// 服务上下文可被多个 goroutine 访问；具体 add-in 的并发约束由各自接口说明。
type IService interface {
	service.Context
	// AppConf 返回合并后的应用配置。
	AppConf() *viper.Viper
	// ServiceConf 返回以服务名为键的配置子树；不存在时可能返回 nil。
	ServiceConf() *viper.Viper
	// Registry 返回服务发现 add-in；未安装时会 panic。
	Registry() discovery.IRegistry
	// Broker 返回消息代理 add-in；未安装时会 panic。
	Broker() broker.IBroker
	// DistSync 返回分布式同步 add-in；未安装时会 panic。
	DistSync() dsync.IDistSync
	// DistService 返回分布式服务 add-in；未安装时会 panic。
	DistService() dsvc.IDistService
	// DistEntityQuerier 返回分布式实体查询 add-in；未安装时会 panic。
	DistEntityQuerier() dent.IDistEntityQuerier
	// RPC 返回 RPC add-in；未安装时会 panic。
	RPC() rpc.IRPC
	// ReplicaNo 返回当前服务在本次应用启动中的副本序号，从 0 开始。
	ReplicaNo() int
	// Memory 返回服务私有的并发键值存储。
	Memory() *sync.Map
	// BuildRuntime 创建绑定当前服务的运行时构建器。
	BuildRuntime() *RuntimeCreator
	// BuildEntityPT 创建绑定当前服务及 prototype 名称的实体原型构建器。
	BuildEntityPT(prototype string) *EntityPTCreator
	// BuildEntity 创建绑定当前服务及 prototype 名称的实体构建器。
	BuildEntity(prototype string) *EntityCreator
	// L 返回当前服务的结构化日志器。
	L() *zap.Logger
	// S 返回当前服务的 SugaredLogger。
	S() *zap.SugaredLogger
}

type iService interface {
	getStarted() *atomic.Bool
	getRuntimeAssembler() *RuntimeAssembler
}

// ServiceBehavior 提供 IService 的默认实现，供自定义服务匿名嵌入。
type ServiceBehavior struct {
	service.ContextBehavior
	started          atomic.Bool
	memory           sync.Map
	runtimeAssembler RuntimeAssembler
}

// AppConf 返回合并后的应用配置。
// 服务启动前直接读取装配配置，启动后通过 conf add-in 读取。
func (svc *ServiceBehavior) AppConf() *viper.Viper {
	if !svc.started.Load() {
		return svc.getConf()
	}
	return addins.Conf.Require(svc).AppConf()
}

// ServiceConf 返回以当前服务名为键的配置子树；不存在时可能返回 nil。
func (svc *ServiceBehavior) ServiceConf() *viper.Viper {
	if !svc.started.Load() {
		return svc.getConf().Sub(svc.Name())
	}
	return addins.Conf.Require(svc).ServiceConf()
}

// Registry 返回服务发现 add-in；未安装时会 panic。
func (svc *ServiceBehavior) Registry() discovery.IRegistry {
	return addins.Discovery.Require(svc)
}

// Broker 返回消息代理 add-in；未安装时会 panic。
func (svc *ServiceBehavior) Broker() broker.IBroker {
	return addins.Broker.Require(svc)
}

// DistSync 返回分布式同步 add-in；未安装时会 panic。
func (svc *ServiceBehavior) DistSync() dsync.IDistSync {
	return addins.Dsync.Require(svc)
}

// DistService 返回分布式服务 add-in；未安装时会 panic。
func (svc *ServiceBehavior) DistService() dsvc.IDistService {
	return addins.Dsvc.Require(svc)
}

// DistEntityQuerier 返回分布式实体查询 add-in；未安装时会 panic。
func (svc *ServiceBehavior) DistEntityQuerier() dent.IDistEntityQuerier {
	return addins.Dentq.Require(svc)
}

// RPC 返回 RPC add-in；未安装时会 panic。
func (svc *ServiceBehavior) RPC() rpc.IRPC {
	return addins.RPC.Require(svc)
}

// ReplicaNo 返回当前服务在本次应用启动中的副本序号，从 0 开始。
func (svc *ServiceBehavior) ReplicaNo() int {
	v, _ := svc.Memory().Load(memReplicaNo)
	startupNo, ok := v.(int)
	if !ok {
		exception.Panicf("%w: service memory %q not exists", ErrFramework, memReplicaNo)
	}
	return startupNo
}

// Memory 返回服务私有的并发键值存储。
func (svc *ServiceBehavior) Memory() *sync.Map {
	return &svc.memory
}

// BuildRuntime 创建绑定当前服务的运行时构建器。
func (svc *ServiceBehavior) BuildRuntime() *RuntimeCreator {
	return BuildRuntime(reinterpret.Cast[IService](service.UnsafeContext(svc).Instance()))
}

// BuildEntityPT 创建绑定当前服务及 prototype 名称的实体原型构建器。
func (svc *ServiceBehavior) BuildEntityPT(prototype string) *EntityPTCreator {
	return BuildEntityPT(service.UnsafeContext(svc).Instance(), prototype)
}

// BuildEntity 创建绑定当前服务及 prototype 名称的实体构建器。
func (svc *ServiceBehavior) BuildEntity(prototype string) *EntityCreator {
	return BuildEntity(reinterpret.Cast[IService](service.UnsafeContext(svc).Instance()), prototype).SetRuntimeCreator(svc.BuildRuntime())
}

// L 返回当前服务的结构化日志器。
func (svc *ServiceBehavior) L() *zap.Logger {
	if !svc.started.Load() {
		return svc.getLogger()
	}
	return log.L(svc)
}

// S 返回当前服务的 SugaredLogger。
func (svc *ServiceBehavior) S() *zap.SugaredLogger {
	if !svc.started.Load() {
		return svc.getLogger().Sugar()
	}
	return log.S(svc)
}

func (svc *ServiceBehavior) getStarted() *atomic.Bool {
	return &svc.started
}

func (svc *ServiceBehavior) getRuntimeAssembler() *RuntimeAssembler {
	return &svc.runtimeAssembler
}

func (svc *ServiceBehavior) getConf() *viper.Viper {
	v, _ := svc.Memory().Load(memConf)
	conf, ok := v.(*viper.Viper)
	if !ok {
		exception.Panicf("%w: service memory %q not exists", ErrFramework, memConf)
	}
	return conf
}

func (svc *ServiceBehavior) getLogger() *zap.Logger {
	v, _ := svc.Memory().Load(memLogger)
	logger, ok := v.(*zap.Logger)
	if !ok {
		exception.Panicf("%w: service memory %q not exists", ErrFramework, memLogger)
	}
	return logger
}
