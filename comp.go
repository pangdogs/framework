package framework

import (
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/plugin"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
)

// ComponentBehavior 组件行为，在开发新组件时，匿名嵌入至组件结构体中
type ComponentBehavior struct {
	ec.ComponentBehavior
}

// GetRuntime 获取运行时
func (c *ComponentBehavior) GetRuntime() Runtime {
	return Runtime{Ctx: runtime.Current(c)}
}

// GetService 获取服务
func (c *ComponentBehavior) GetService() Service {
	return Service{Ctx: service.Current(c)}
}

// GetPluginBundle 获取插件包
func (c *ComponentBehavior) GetPluginBundle() plugin.PluginBundle {
	return runtime.Current(c).GetPluginBundle()
}

// IsAlive 是否活跃
func (c *ComponentBehavior) IsAlive() bool {
	return c.GetState() <= ec.ComponentState_Alive
}
