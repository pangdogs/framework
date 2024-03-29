package framework

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
)

// CreateRuntime 创建运行时
func CreateRuntime(ctx service.Context) RuntimeCreator {
	if ctx == nil {
		panic(fmt.Errorf("%w: ctx is nil", core.ErrArgs))
	}
	return RuntimeCreator{
		servCtx: ctx,
		runtime: nil,
		settings: _RuntimeSettings{
			Name:                 "",
			AutoRecover:          ctx.GetAutoRecover(),
			ReportError:          ctx.GetReportError(),
			FPS:                  0,
			AutoRun:              true,
			ProcessQueueCapacity: 128,
		},
	}
}

// RuntimeCreator 运行时构建器
type RuntimeCreator struct {
	servCtx  service.Context
	runtime  _IRuntime
	settings _RuntimeSettings
}

// Setup 安装运行时
func (c RuntimeCreator) Setup(rt any) RuntimeCreator {
	if c.servCtx == nil {
		panic("setting servCtx is nil")
	}

	if rt == nil {
		panic(fmt.Errorf("%w: rt is nil", core.ErrArgs))
	}

	_rt, ok := rt.(_IRuntime)
	if !ok {
		panic(fmt.Errorf("%w: incorrect rt type", core.ErrArgs))
	}

	c.runtime = _rt
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

// AutoRun 自动开始运行
func (c RuntimeCreator) AutoRun(auto bool) RuntimeCreator {
	c.settings.AutoRun = auto
	return c
}

// ProcessQueueCapacity 任务处理流水线大小
func (c RuntimeCreator) ProcessQueueCapacity(cap int) RuntimeCreator {
	c.settings.ProcessQueueCapacity = cap
	return c
}

// Spawn 创建运行时
func (c RuntimeCreator) Spawn() core.Runtime {
	if c.servCtx == nil {
		panic("setting servCtx is nil")
	}

	rt := c.runtime
	if rt == nil {
		rt = &RuntimeBehavior{}
	}
	rt.setup(c.servCtx, rt)

	return rt.generate(c.settings)
}
