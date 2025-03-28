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
)

// GetService 获取服务实例
func GetService(provider runtime.ConcurrentContextProvider) IService {
	return reinterpret.Cast[IService](service.Current(provider))
}

// IService 服务实例接口
type IService interface {
	service.Context
	// GetConf 获取配置插件
	GetConf() conf.IConfig
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
	// GetStartupConf 获取启动参数配置
	GetStartupConf() *viper.Viper
	// GetMemKV 获取服务内存KV数据库
	GetMemKV() *sync.Map
	// BuildRuntime 创建运行时
	BuildRuntime() RuntimeCreator
	// BuildEntityPT 创建实体原型
	BuildEntityPT(prototype string) core.EntityPTCreator
	// BuildEntityAsync 创建实体
	BuildEntityAsync(prototype string) EntityCreatorAsync
}

// Service 服务实例
type Service struct {
	service.ContextBehavior
	runtimeGeneric RuntimeGeneric
}

// GetConf 获取配置插件
func (inst *Service) GetConf() conf.IConfig {
	return conf.Using(inst)
}

// GetRegistry 获取服务发现插件
func (inst *Service) GetRegistry() discovery.IRegistry {
	return discovery.Using(inst)
}

// GetBroker 获取消息队列中间件插件
func (inst *Service) GetBroker() broker.IBroker {
	return broker.Using(inst)
}

// GetDistSync 获取分布式同步插件
func (inst *Service) GetDistSync() dsync.IDistSync {
	return dsync.Using(inst)
}

// GetDistService 获取分布式服务插件
func (inst *Service) GetDistService() dsvc.IDistService {
	return dsvc.Using(inst)
}

// GetDistEntityQuerier 获取分布式实体查询插件
func (inst *Service) GetDistEntityQuerier() dentq.IDistEntityQuerier {
	return dentq.Using(inst)
}

// GetRPC 获取RPC支持插件
func (inst *Service) GetRPC() rpc.IRPC {
	return rpc.Using(inst)
}

// GetStartupNo 获取启动序号
func (inst *Service) GetStartupNo() int {
	v, _ := inst.GetMemKV().Load("startup.no")
	if v == nil {
		exception.Panicf("%w: service memory kv startup.no not existed", ErrFramework)
	}
	return v.(int)
}

// GetStartupConf 获取启动参数配置
func (inst *Service) GetStartupConf() *viper.Viper {
	v, _ := inst.GetMemKV().Load("startup.conf")
	if v == nil {
		exception.Panicf("%w: service memory kv startup.conf not existed", ErrFramework)
	}
	return v.(*viper.Viper)
}

// GetMemKV 获取服务内存KV数据库
func (inst *Service) GetMemKV() *sync.Map {
	memKV, _ := inst.Value("mem_kv").(*sync.Map)
	if memKV == nil {
		exception.Panicf("%w: service memory not existed", ErrFramework)
	}
	return memKV
}

// BuildRuntime 创建运行时
func (inst *Service) BuildRuntime() RuntimeCreator {
	return BuildRuntime(service.UnsafeContext(inst).GetOptions().InstanceFace.Iface).Setup(&inst.runtimeGeneric)
}

// BuildEntityPT 创建实体原型
func (inst *Service) BuildEntityPT(prototype string) core.EntityPTCreator {
	return core.BuildEntityPT(service.UnsafeContext(inst).GetOptions().InstanceFace.Iface, prototype)
}

// BuildEntityAsync 创建实体
func (inst *Service) BuildEntityAsync(prototype string) EntityCreatorAsync {
	return BuildEntityAsync(service.UnsafeContext(inst).GetOptions().InstanceFace.Iface, prototype).SetRuntimeCreator(inst.BuildRuntime())
}
