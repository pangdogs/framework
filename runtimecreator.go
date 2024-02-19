package framework

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
)

func CreateRuntime(ctx service.Context) RuntimeCreator {
	if ctx == nil {
		panic(fmt.Errorf("%w: ctx is nil", core.ErrArgs))
	}
	return RuntimeCreator{
		context: ctx,
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

type RuntimeCreator struct {
	context  service.Context
	runtime  _IRuntime
	settings _RuntimeSettings
}

func (c RuntimeCreator) Setup(rt any) RuntimeCreator {
	if c.context == nil {
		panic("setting context is nil")
	}

	if rt == nil {
		panic(fmt.Errorf("%w: rt is nil", core.ErrArgs))
	}

	_rt, ok := rt.(_IRuntime)
	if !ok {
		panic(fmt.Errorf("%w: incorrect rt type", core.ErrArgs))
	}

	c.runtime = _rt
	c.runtime.init(c.context, _rt)
	return c
}

func (c RuntimeCreator) Name(name string) RuntimeCreator {
	c.settings.Name = name
	return c
}

func (c RuntimeCreator) PanicHandling(autoRecover bool, reportError chan error) RuntimeCreator {
	c.settings.AutoRecover = autoRecover
	c.settings.ReportError = reportError
	return c
}

func (c RuntimeCreator) FPS(fps float32) RuntimeCreator {
	c.settings.FPS = fps
	return c
}

func (c RuntimeCreator) AutoRun(auto bool) RuntimeCreator {
	c.settings.AutoRun = auto
	return c
}

func (c RuntimeCreator) ProcessQueueCapacity(cap int) RuntimeCreator {
	c.settings.ProcessQueueCapacity = cap
	return c
}

func (c RuntimeCreator) Spawn() core.Runtime {
	if c.context == nil {
		panic("setting context is nil")
	}

	rt := c.runtime

	if rt == nil {
		rt = &RuntimeBehavior{}
		rt.init(c.context, c.runtime)
	}

	return rt.generate(c.settings)
}
