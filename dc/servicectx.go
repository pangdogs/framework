package dc

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
	"git.golaxy.org/framework"
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

// ServiceCtx 服务上下文
type ServiceCtx struct {
	service.Context
}

// GetConf 获取配置插件
func (ctx ServiceCtx) GetConf() conf.IVisitConf {
	return conf.Using(ctx.Context)
}

// GetRegistry 获取服务发现插件
func (ctx ServiceCtx) GetRegistry() discovery.IRegistry {
	return discovery.Using(ctx.Context)
}

// GetBroker 获取broker插件
func (ctx ServiceCtx) GetBroker() broker.IBroker {
	return broker.Using(ctx.Context)
}

// GetDistSync 获取分布式同步插件
func (ctx ServiceCtx) GetDistSync() dsync.IDistSync {
	return dsync.Using(ctx.Context)
}

// GetDistService 获取分布式服务插件
func (ctx ServiceCtx) GetDistService() dserv.IDistService {
	return dserv.Using(ctx.Context)
}

// GetDistEntityQuerier 获取分布式实体查询插件
func (ctx ServiceCtx) GetDistEntityQuerier() dentq.IDistEntityQuerier {
	return dentq.Using(ctx.Context)
}

// GetRPC 获取RPC支持插件
func (ctx ServiceCtx) GetRPC() rpc.IRPC {
	return rpc.Using(ctx.Context)
}

// GetStartupConf 获取启动参数配置
func (ctx ServiceCtx) GetStartupConf() *viper.Viper {
	v, _ := ctx.GetMemKVs().Load("startup.conf")
	if v == nil {
		panic("service memory startup.conf not existed")
	}
	return v.(*viper.Viper)
}

// GetMemKVs 获取服务内存KV数据库
func (ctx ServiceCtx) GetMemKVs() *sync.Map {
	memKVs, _ := ctx.Value("mem_kvs").(*sync.Map)
	if memKVs == nil {
		panic("service memory not existed")
	}
	return memKVs
}

// CreateRuntime 创建运行时
func (ctx ServiceCtx) CreateRuntime() framework.RuntimeCreator {
	return framework.CreateRuntime(ctx.Context)
}

// CreateEntityPT 创建实体原型
func (ctx ServiceCtx) CreateEntityPT() core.EntityPTCreator {
	return core.CreateEntityPT(ctx.Context)
}
