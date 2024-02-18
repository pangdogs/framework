package framework

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
)

func CreateRuntime(ctx service.Context) *RuntimeCreator {
	if ctx == nil {
		panic(fmt.Errorf("%w: ctx is nil", core.ErrArgs))
	}
	return &RuntimeCreator{
		servCtx:              ctx,
		runtime:              nil,
		name:                 "",
		autoRecover:          false,
		reportError:          nil,
		frameFPS:             0,
		frameBlink:           false,
		autoRun:              true,
		processQueueCapacity: 128,
	}
}

type RuntimeCreator struct {
	servCtx              service.Context
	runtime              _IRuntime
	name                 string
	autoRecover          bool
	reportError          chan error
	frameFPS             float32
	frameBlink           bool
	autoRun              bool
	processQueueCapacity int
}

func (c *RuntimeCreator) Setup(rt any) *RuntimeCreator {
	c.runtime = rt.(_IRuntime)
	return c
}

func (c *RuntimeCreator) Name(name string) *RuntimeCreator {
	c.name = name
	return c
}

func (c *RuntimeCreator) ErrorHandling(autoRecover bool, reportError chan error) *RuntimeCreator {
	c.autoRecover = autoRecover
	c.reportError = reportError
	return c
}

func (c *RuntimeCreator) Frames(fps float32, blink bool) *RuntimeCreator {
	c.frameFPS = fps
	c.frameBlink = blink
	return c
}

func (c *RuntimeCreator) AutoRun(auto bool) *RuntimeCreator {
	c.autoRun = auto
	return c
}

func (c *RuntimeCreator) ProcessQueueCapacity(cap int) *RuntimeCreator {
	c.processQueueCapacity = cap
	return c
}

func (c *RuntimeCreator) Spawn() core.Runtime {
	var rt _IRuntime

	if c.runtime != nil {
		rt = c.runtime
	} else {
		rt = &RuntimeBehavior{}
	}

	rt.init(c, rt)

	return rt.generate()
}
