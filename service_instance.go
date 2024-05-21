package framework

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
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
	Instantiation() service.Context
}

// ServiceInstance 服务实例
type ServiceInstance struct {
	service.ContextBehavior
}

// GetConf 获取配置插件
func (instance *ServiceInstance) GetConf() conf.IConfig {
	return conf.Using(instance)
}

// GetRegistry 获取服务发现插件
func (instance *ServiceInstance) GetRegistry() discovery.IRegistry {
	return discovery.Using(instance)
}

// GetBroker 获取broker插件
func (instance *ServiceInstance) GetBroker() broker.IBroker {
	return broker.Using(instance)
}

// GetDistSync 获取分布式同步插件
func (instance *ServiceInstance) GetDistSync() dsync.IDistSync {
	return dsync.Using(instance)
}

// GetDistService 获取分布式服务插件
func (instance *ServiceInstance) GetDistService() dserv.IDistService {
	return dserv.Using(instance)
}

// GetDistEntityQuerier 获取分布式实体查询插件
func (instance *ServiceInstance) GetDistEntityQuerier() dentq.IDistEntityQuerier {
	return dentq.Using(instance)
}

// GetRPC 获取RPC支持插件
func (instance *ServiceInstance) GetRPC() rpc.IRPC {
	return rpc.Using(instance)
}

// GetStartupIdx 获取启动索引
func (instance *ServiceInstance) GetStartupIdx() int {
	v, _ := instance.GetMemKVs().Load("startup.idx")
	if v == nil {
		panic("service memory startup.idx not existed")
	}
	return v.(int)
}

// GetStartupConf 获取启动参数配置
func (instance *ServiceInstance) GetStartupConf() *viper.Viper {
	v, _ := instance.GetMemKVs().Load("startup.conf")
	if v == nil {
		panic("service memory startup.conf not existed")
	}
	return v.(*viper.Viper)
}

// GetMemKVs 获取服务内存KV数据库
func (instance *ServiceInstance) GetMemKVs() *sync.Map {
	memKVs, _ := instance.Value("mem_kvs").(*sync.Map)
	if memKVs == nil {
		panic("service memory not existed")
	}
	return memKVs
}

// CreateRuntime 创建运行时
func (instance *ServiceInstance) CreateRuntime() RuntimeCreator {
	return CreateRuntime(instance)
}

// CreateEntityPT 创建实体原型
func (instance *ServiceInstance) CreateEntityPT() core.EntityPTCreator {
	return core.CreateEntityPT(instance)
}
