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
	"git.golaxy.org/framework/plugins/broker"
	"git.golaxy.org/framework/plugins/conf"
	"git.golaxy.org/framework/plugins/dentq"
	"git.golaxy.org/framework/plugins/discovery"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/dsync"
	"git.golaxy.org/framework/plugins/rpc"
	"github.com/spf13/viper"
	"sync"
)

// IServiceInstantiation 服务实例化接口
type IServiceInstantiation interface {
	Instantiation() IServiceInstance
}

// GetServiceInstance 获取服务实例
func GetServiceInstance(provider runtime.ConcurrentContextProvider) IServiceInstance {
	return reinterpret.Cast[IServiceInstance](service.Current(provider))
}

// IServiceInstance 服务实例接口
type IServiceInstance interface {
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
	GetDistService() dserv.IDistService
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
	// CreateRuntime 创建运行时
	CreateRuntime() RuntimeCreator
	// CreateEntityPT 创建实体原型
	CreateEntityPT(prototype string) core.EntityPTCreator
	// CreateConcurrentEntity 创建实体
	CreateConcurrentEntity(prototype string) ConcurrentEntityCreator
}

// ServiceInstance 服务实例
type ServiceInstance struct {
	service.ContextBehavior
}

// GetConf 获取配置插件
func (inst *ServiceInstance) GetConf() conf.IConfig {
	return conf.Using(inst)
}

// GetRegistry 获取服务发现插件
func (inst *ServiceInstance) GetRegistry() discovery.IRegistry {
	return discovery.Using(inst)
}

// GetBroker 获取消息队列中间件插件
func (inst *ServiceInstance) GetBroker() broker.IBroker {
	return broker.Using(inst)
}

// GetDistSync 获取分布式同步插件
func (inst *ServiceInstance) GetDistSync() dsync.IDistSync {
	return dsync.Using(inst)
}

// GetDistService 获取分布式服务插件
func (inst *ServiceInstance) GetDistService() dserv.IDistService {
	return dserv.Using(inst)
}

// GetDistEntityQuerier 获取分布式实体查询插件
func (inst *ServiceInstance) GetDistEntityQuerier() dentq.IDistEntityQuerier {
	return dentq.Using(inst)
}

// GetRPC 获取RPC支持插件
func (inst *ServiceInstance) GetRPC() rpc.IRPC {
	return rpc.Using(inst)
}

// GetStartupNo 获取启动序号
func (inst *ServiceInstance) GetStartupNo() int {
	v, _ := inst.GetMemKV().Load("startup.no")
	if v == nil {
		panic("service memory kv startup.no not existed")
	}
	return v.(int)
}

// GetStartupConf 获取启动参数配置
func (inst *ServiceInstance) GetStartupConf() *viper.Viper {
	v, _ := inst.GetMemKV().Load("startup.conf")
	if v == nil {
		panic("service memory kv startup.conf not existed")
	}
	return v.(*viper.Viper)
}

// GetMemKV 获取服务内存KV数据库
func (inst *ServiceInstance) GetMemKV() *sync.Map {
	memKV, _ := inst.Value("mem_kv").(*sync.Map)
	if memKV == nil {
		panic("service memory not existed")
	}
	return memKV
}

// CreateRuntime 创建运行时
func (inst *ServiceInstance) CreateRuntime() RuntimeCreator {
	return CreateRuntime(service.UnsafeContext(inst).GetOptions().CompositeFace.Iface)
}

// CreateEntityPT 创建实体原型
func (inst *ServiceInstance) CreateEntityPT(prototype string) core.EntityPTCreator {
	return core.CreateEntityPT(service.UnsafeContext(inst).GetOptions().CompositeFace.Iface, prototype)
}

// CreateConcurrentEntity 创建实体
func (inst *ServiceInstance) CreateConcurrentEntity(prototype string) ConcurrentEntityCreator {
	return CreateConcurrentEntity(service.UnsafeContext(inst).GetOptions().CompositeFace.Iface, prototype)
}
