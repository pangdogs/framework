package oc

import (
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

type ServiceCtx struct {
	service.Context
}

func (ctx ServiceCtx) GetConf() conf.IVisitConf {
	return conf.Using(ctx.Context)
}

func (ctx ServiceCtx) GetRegistry() discovery.IRegistry {
	return discovery.Using(ctx.Context)
}

func (ctx ServiceCtx) GetBroker() broker.IBroker {
	return broker.Using(ctx.Context)
}

func (ctx ServiceCtx) GetDistSync() dsync.IDistSync {
	return dsync.Using(ctx.Context)
}

func (ctx ServiceCtx) GetDistService() dserv.IDistService {
	return dserv.Using(ctx.Context)
}

func (ctx ServiceCtx) GetDistEntityQuerier() dentq.IDistEntityQuerier {
	return dentq.Using(ctx.Context)
}

func (ctx ServiceCtx) GetRPC() rpc.IRPC {
	return rpc.Using(ctx.Context)
}

func (ctx ServiceCtx) GetStartupConf() *viper.Viper {
	v, _ := ctx.GetMemKVs().Load("startup.conf")
	if v == nil {
		panic("service memory startup.conf not existed")
	}
	return v.(*viper.Viper)
}

func (ctx ServiceCtx) GetMemKVs() *sync.Map {
	memKVs, _ := ctx.Value("mem_kvs").(*sync.Map)
	if memKVs == nil {
		panic("service memory not existed")
	}
	return memKVs
}
