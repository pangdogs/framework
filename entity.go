package framework

import (
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/plugin"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
)

// EntityBehavior 实体行为，在需要扩展实体能力时，匿名嵌入至实体结构体中
type EntityBehavior struct {
	ec.EntityBehavior
}

// GetRuntime 获取运行时
func (e *EntityBehavior) GetRuntime() Runtime {
	return Runtime{Ctx: runtime.Current(e)}
}

// GetService 获取服务
func (e *EntityBehavior) GetService() Service {
	return Service{Ctx: service.Current(e)}
}

// GetPluginBundle 获取插件包
func (e *EntityBehavior) GetPluginBundle() plugin.PluginBundle {
	return runtime.Current(e).GetPluginBundle()
}

// IsAlive 是否活跃
func (e *EntityBehavior) IsAlive() bool {
	return e.GetState() <= ec.EntityState_Alive
}
