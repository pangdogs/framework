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
	"sync"
)

// Service 服务
type Service struct {
	Ctx service.Context
}

// GetConf 获取配置插件
func (serv Service) GetConf() conf.IConfig {
	return conf.Using(serv.Ctx)
}

// GetRegistry 获取服务发现插件
func (serv Service) GetRegistry() discovery.IRegistry {
	return discovery.Using(serv.Ctx)
}

// GetBroker 获取broker插件
func (serv Service) GetBroker() broker.IBroker {
	return broker.Using(serv.Ctx)
}

// GetDistSync 获取分布式同步插件
func (serv Service) GetDistSync() dsync.IDistSync {
	return dsync.Using(serv.Ctx)
}

// GetDistService 获取分布式服务插件
func (serv Service) GetDistService() dserv.IDistService {
	return dserv.Using(serv.Ctx)
}

// GetDistEntityQuerier 获取分布式实体查询插件
func (serv Service) GetDistEntityQuerier() dentq.IDistEntityQuerier {
	return dentq.Using(serv.Ctx)
}

// GetRPC 获取RPC支持插件
func (serv Service) GetRPC() rpc.IRPC {
	return rpc.Using(serv.Ctx)
}

// GetStartupNo 获取启动序号
func (serv Service) GetStartupNo() int {
	v, _ := serv.GetMemKV().Load("startup.no")
	if v == nil {
		panic("service memory kv startup.no not existed")
	}
	return v.(int)
}

// GetMemKV 获取服务内存KV数据库
func (serv Service) GetMemKV() *sync.Map {
	memKV, _ := serv.Ctx.Value("mem_kv").(*sync.Map)
	if memKV == nil {
		panic("service memory not existed")
	}
	return memKV
}

// CreateRuntime 创建运行时
func (serv Service) CreateRuntime() RuntimeCreator {
	return CreateRuntime(serv.Ctx)
}

// CreateEntityPT 创建实体原型
func (serv Service) CreateEntityPT(prototype string) core.EntityPTCreator {
	return core.CreateEntityPT(serv.Ctx, prototype)
}
