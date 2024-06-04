package framework

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/reinterpret"
)

// CreateRuntime 创建运行时
func CreateRuntime(ctx service.Context) RuntimeCreator {
	if ctx == nil {
		panic(fmt.Errorf("%w: ctx is nil", core.ErrArgs))
	}
	return RuntimeCreator{
		servCtx: ctx,
		generic: nil,
		settings: _RuntimeSettings{
			Name:                 "",
			AutoRecover:          ctx.GetAutoRecover(),
			ReportError:          ctx.GetReportError(),
			FPS:                  0,
			ProcessQueueCapacity: 128,
		},
	}
}

// RuntimeCreator 运行时构建器
type RuntimeCreator struct {
	servCtx  service.Context
	generic  iRuntimeGeneric
	settings _RuntimeSettings
}

// Setup 安装运行时泛化类型
func (c RuntimeCreator) Setup(generic any) RuntimeCreator {
	if c.servCtx == nil {
		panic("setting servCtx is nil")
	}

	if generic == nil {
		panic(fmt.Errorf("%w: generic is nil", core.ErrArgs))
	}

	_generic, ok := generic.(iRuntimeGeneric)
	if !ok {
		panic(fmt.Errorf("%w: incorrect generic type", core.ErrArgs))
	}

	c.generic = _generic
	return c
}

// Name 名称
func (c RuntimeCreator) Name(name string) RuntimeCreator {
	c.settings.Name = name
	return c
}

// PanicHandling panic时的处理方式
func (c RuntimeCreator) PanicHandling(autoRecover bool, reportError chan error) RuntimeCreator {
	c.settings.AutoRecover = autoRecover
	c.settings.ReportError = reportError
	return c
}

// FPS 帧率
func (c RuntimeCreator) FPS(fps float32) RuntimeCreator {
	c.settings.FPS = fps
	return c
}

// ProcessQueueCapacity 任务处理流水线大小
func (c RuntimeCreator) ProcessQueueCapacity(cap int) RuntimeCreator {
	c.settings.ProcessQueueCapacity = cap
	return c
}

// Spawn 创建运行时
func (c RuntimeCreator) Spawn() IRuntimeInstance {
	if c.servCtx == nil {
		panic("setting servCtx is nil")
	}

	generic := c.generic
	if generic == nil {
		generic = &RuntimeGenericT[RuntimeInstance]{}
	}
	generic.init(c.servCtx, generic)

	return reinterpret.Cast[IRuntimeInstance](runtime.Current(generic.generate(c.settings)))
}
