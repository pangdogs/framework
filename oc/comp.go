package oc

import (
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
)

// ComponentBehavior 组件行为，需要在开发新组件时，匿名嵌入至组件结构体中
type ComponentBehavior struct {
	ec.ComponentBehavior
}

func (c *ComponentBehavior) GetRuntimeCtx() runtime.Context {
	return runtime.Current(c)
}

func (c *ComponentBehavior) GetServiceCtx() service.Context {
	return service.Current(c)
}