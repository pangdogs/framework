package framework

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
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

// GetMemKV 获取服务内存KV数据库
func (inst *RuntimeInstance) GetMemKV() *sync.Map {
	memKV, _ := service.Current(inst).Value("mem_kv").(*sync.Map)
	if memKV == nil {
		panic("service memory not existed")
	}
	return memKV
}

// CreateEntity 创建实体
func (inst *RuntimeInstance) CreateEntity(prototype string) core.EntityCreator {
	return core.CreateEntity(runtime.Current(inst), prototype)
}
