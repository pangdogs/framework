package framework

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"github.com/spf13/viper"
	"sync"
)

// IRuntimeInstantiation 运行时实例化接口
type IRuntimeInstantiation interface {
	Instantiation() runtime.Context
}

// RuntimeInstance 运行时实例
type RuntimeInstance struct {
	runtime.ContextBehavior
}

// GetStartupConf 获取启动参数配置
func (inst *RuntimeInstance) GetStartupConf() *viper.Viper {
	v, _ := inst.GetMemKV().Load("startup.conf")
	if v == nil {
		panic("service memory kv startup.conf not existed")
	}
	return v.(*viper.Viper)
}

// GetMemKV 获取服务内存KV数据库
func (inst *RuntimeInstance) GetMemKV() *sync.Map {
	memKV, _ := service.Current(inst).Value("mem_kv").(*sync.Map)
	if memKV == nil {
		panic("service memory not existed")
	}
	return memKV
}

// CreateEntity 创建实体
func (inst *RuntimeInstance) CreateEntity() core.EntityCreator {
	return core.CreateEntity(inst)
}
