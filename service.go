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

// GetService 获取服务实例
func GetService(provider runtime.ConcurrentContextProvider) IService {
	return reinterpret.Cast[IService](service.Current(provider))
}

// IService 服务实例接口
type IService interface {
	service.Context
	// AppConf 获取当前应用程序配置
	AppConf() *viper.Viper
	// ServiceConf 获取当前服务配置
	ServiceConf() *viper.Viper
	// Registry 获取服务发现插件
	Registry() discovery.IRegistry
	// Broker 获取消息队列中间件插件
	Broker() broker.IBroker
	// DistSync 获取分布式同步插件
	DistSync() dsync.IDistSync
	// DistService 获取分布式服务插件
	DistService() dsvc.IDistService
	// DistEntityQuerier 获取分布式实体查询插件
	DistEntityQuerier() dent.IDistEntityQuerier
	// RPC 获取RPC支持插件
	RPC() rpc.IRPC
	// ReplicaNo 获取副本序号
	ReplicaNo() int
	// Memory 获取服务内存KV存储
	Memory() *sync.Map
	// BuildRuntime 创建运行时
	BuildRuntime() *RuntimeCreator
	// BuildEntityPT 创建实体原型
	BuildEntityPT(prototype string) *EntityPTCreator
	// BuildEntity 创建实体
	BuildEntity(prototype string) *EntityCreator
	// L 结构化日志
	L() *zap.Logger
	// S 传统日志
	S() *zap.SugaredLogger
}

type iService interface {
	getStarted() *atomic.Bool
	getRuntimeAssembler() *RuntimeAssembler
}

// ServiceBehavior 服务实例行为
type ServiceBehavior struct {
	service.ContextBehavior
	started          atomic.Bool
	memory           sync.Map
	runtimeAssembler RuntimeAssembler
}

// AppConf 获取当前应用配置
func (svc *ServiceBehavior) AppConf() *viper.Viper {
	if !svc.started.Load() {
		return svc.getConf()
	}
	return addins.Conf.Require(svc).AppConf()
}

// ServiceConf 获取当前服务配置
func (svc *ServiceBehavior) ServiceConf() *viper.Viper {
	if !svc.started.Load() {
		return svc.getConf().Sub(svc.Name())
	}
	return addins.Conf.Require(svc).ServiceConf()
}

// Registry 获取服务发现插件
func (svc *ServiceBehavior) Registry() discovery.IRegistry {
	return addins.Discovery.Require(svc)
}

// Broker 获取消息队列中间件插件
func (svc *ServiceBehavior) Broker() broker.IBroker {
	return addins.Broker.Require(svc)
}

// DistSync 获取分布式同步插件
func (svc *ServiceBehavior) DistSync() dsync.IDistSync {
	return addins.Dsync.Require(svc)
}

// DistService 获取分布式服务插件
func (svc *ServiceBehavior) DistService() dsvc.IDistService {
	return addins.Dsvc.Require(svc)
}

// DistEntityQuerier 获取分布式实体查询插件
func (svc *ServiceBehavior) DistEntityQuerier() dent.IDistEntityQuerier {
	return addins.Dentq.Require(svc)
}

// RPC 获取RPC支持插件
func (svc *ServiceBehavior) RPC() rpc.IRPC {
	return addins.RPC.Require(svc)
}

// ReplicaNo 获取副本序号
func (svc *ServiceBehavior) ReplicaNo() int {
	v, _ := svc.Memory().Load(memReplicaNo)
	startupNo, ok := v.(int)
	if !ok {
		exception.Panicf("%w: service memory %q not exists", ErrFramework, memReplicaNo)
	}
	return startupNo
}

// Memory 获取服务内存KV存储
func (svc *ServiceBehavior) Memory() *sync.Map {
	return &svc.memory
}

// BuildRuntime 创建运行时
func (svc *ServiceBehavior) BuildRuntime() *RuntimeCreator {
	rtCtor := BuildRuntime(reinterpret.Cast[IService](service.UnsafeContext(svc).Instance()))
	rtCtor.assembler = &svc.runtimeAssembler
	return rtCtor
}

// BuildEntityPT 创建实体原型
func (svc *ServiceBehavior) BuildEntityPT(prototype string) *EntityPTCreator {
	return BuildEntityPT(service.UnsafeContext(svc).Instance(), prototype)
}

// BuildEntity 创建实体
func (svc *ServiceBehavior) BuildEntity(prototype string) *EntityCreator {
	return BuildEntity(reinterpret.Cast[IService](service.UnsafeContext(svc).Instance()), prototype).SetRuntimeCreator(svc.BuildRuntime())
}

// L 结构化日志
func (svc *ServiceBehavior) L() *zap.Logger {
	return log.L(svc)
}

// S 传统日志
func (svc *ServiceBehavior) S() *zap.SugaredLogger {
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
