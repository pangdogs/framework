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
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/framework/addins/broker"
	"git.golaxy.org/framework/addins/conf"
	"git.golaxy.org/framework/addins/dentq"
	"git.golaxy.org/framework/addins/discovery"
	"git.golaxy.org/framework/addins/dsvc"
	"git.golaxy.org/framework/addins/dsync"
	"git.golaxy.org/framework/addins/rpc"
	"github.com/spf13/viper"
	"sync"
	"sync/atomic"
)

// GetService 获取服务实例
func GetService(provider runtime.ConcurrentContextProvider) IService {
	return reinterpret.Cast[IService](service.Current(provider))
}

// IService 服务实例接口
type IService interface {
	service.Context
	// GetAppConf 获取当前应用程序配置
	GetAppConf() *viper.Viper
	// GetServiceConf 获取当前服务配置
	GetServiceConf() *viper.Viper
	// GetRegistry 获取服务发现插件
	GetRegistry() discovery.IRegistry
	// GetBroker 获取消息队列中间件插件
	GetBroker() broker.IBroker
	// GetDistSync 获取分布式同步插件
	GetDistSync() dsync.IDistSync
	// GetDistService 获取分布式服务插件
	GetDistService() dsvc.IDistService
	// GetDistEntityQuerier 获取分布式实体查询插件
	GetDistEntityQuerier() dentq.IDistEntityQuerier
	// GetRPC 获取RPC支持插件
	GetRPC() rpc.IRPC
	// GetStartupNo 获取启动序号
	GetStartupNo() int
	// GetMemory 获取服务内存KV存储
	GetMemory() *sync.Map
	// BuildRuntime 创建运行时
	BuildRuntime() *RuntimeCreator
	// BuildEntityPT 创建实体原型
	BuildEntityPT(prototype string) *EntityPTCreator
	// BuildEntityAsync 创建实体
	BuildEntityAsync(prototype string) *EntityCreatorAsync
}

type iService interface {
	getStarted() *atomic.Bool
	getRuntimeGeneric() *RuntimeGeneric
}

// Service 服务实例
type Service struct {
	service.ContextBehavior
	started        atomic.Bool
	memory         sync.Map
	runtimeGeneric RuntimeGeneric
}

// GetAppConf 获取当前应用程序配置
func (svc *Service) GetAppConf() *viper.Viper {
	if !svc.started.Load() {
		return svc.getStartupConf()
	}
	return conf.Using(svc).AppConf()
}

// GetServiceConf 获取当前服务配置
func (svc *Service) GetServiceConf() *viper.Viper {
	if !svc.started.Load() {
		return svc.getStartupConf().Sub(svc.GetName())
	}
	return conf.Using(svc).ServiceConf()
}

// GetRegistry 获取服务发现插件
func (svc *Service) GetRegistry() discovery.IRegistry {
	return discovery.Using(svc)
}

// GetBroker 获取消息队列中间件插件
func (svc *Service) GetBroker() broker.IBroker {
	return broker.Using(svc)
}

// GetDistSync 获取分布式同步插件
func (svc *Service) GetDistSync() dsync.IDistSync {
	return dsync.Using(svc)
}

// GetDistService 获取分布式服务插件
func (svc *Service) GetDistService() dsvc.IDistService {
	return dsvc.Using(svc)
}

// GetDistEntityQuerier 获取分布式实体查询插件
func (svc *Service) GetDistEntityQuerier() dentq.IDistEntityQuerier {
	return dentq.Using(svc)
}

// GetRPC 获取RPC支持插件
func (svc *Service) GetRPC() rpc.IRPC {
	return rpc.Using(svc)
}

// GetStartupNo 获取启动序号
func (svc *Service) GetStartupNo() int {
	v, _ := svc.GetMemory().Load(memStartupNo)
	startupNo, ok := v.(int)
	if !ok {
		exception.Panicf("%w: service memory %q not existed", ErrFramework, memStartupNo)
	}
	return startupNo
}

// GetMemory 获取服务内存KV存储
func (svc *Service) GetMemory() *sync.Map {
	return &svc.memory
}

// BuildRuntime 创建运行时
func (svc *Service) BuildRuntime() *RuntimeCreator {
	rtCtor := BuildRuntime(service.UnsafeContext(svc).GetOptions().InstanceFace.Iface)
	rtCtor.generic = &svc.runtimeGeneric
	return rtCtor
}

// BuildEntityPT 创建实体原型
func (svc *Service) BuildEntityPT(prototype string) *EntityPTCreator {
	return BuildEntityPT(service.UnsafeContext(svc).GetOptions().InstanceFace.Iface, prototype)
}

// BuildEntityAsync 创建实体
func (svc *Service) BuildEntityAsync(prototype string) *EntityCreatorAsync {
	return BuildEntityAsync(service.UnsafeContext(svc).GetOptions().InstanceFace.Iface, prototype).SetRuntimeCreator(svc.BuildRuntime())
}

func (svc *Service) getStarted() *atomic.Bool {
	return &svc.started
}

func (svc *Service) getRuntimeGeneric() *RuntimeGeneric {
	return &svc.runtimeGeneric
}

func (svc *Service) getStartupConf() *viper.Viper {
	v, _ := svc.GetMemory().Load(memStartupConf)
	startupConf, ok := v.(*viper.Viper)
	if !ok {
		exception.Panicf("%w: service memory %q not existed", ErrFramework, memStartupConf)
	}
	return startupConf
}
