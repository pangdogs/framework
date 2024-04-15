package dec

import (
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
)

// ComponentBehavior 组件行为，在开发新组件时，匿名嵌入至组件结构体中
type ComponentBehavior struct {
	ec.ComponentBehavior
}

// GetRuntimeCtx 获取运行时上下文
func (c *ComponentBehavior) GetRuntimeCtx() RuntimeCtx {
	return RuntimeCtx{Context: runtime.Current(c)}
}

// GetServiceCtx 获取服务上下文
func (c *ComponentBehavior) GetServiceCtx() ServiceCtx {
	return ServiceCtx{Context: service.Current(c)}
}
