/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package framework

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/core/utils/uid"
)

// BuildRuntime 创建运行时
func BuildRuntime(svcCtx service.Context) *RuntimeCreator {
	if svcCtx == nil {
		exception.Panicf("%w: %w: svcCtx is nil", ErrFramework, core.ErrArgs)
	}

	return &RuntimeCreator{
		svcCtx: svcCtx,
		settings: _RuntimeSettings{
			Name:                 "",
			PersistId:            uid.Nil,
			AutoRecover:          svcCtx.GetAutoRecover(),
			ReportError:          svcCtx.GetReportError(),
			FPS:                  0,
			ProcessQueueCapacity: 128,
			AutoInjection:        true,
		},
	}
}

// RuntimeCreator 运行时构建器
type RuntimeCreator struct {
	svcCtx   service.Context
	generic  iRuntimeGeneric
	settings _RuntimeSettings
}

// Setup 安装运行时泛化类型
func (c *RuntimeCreator) Setup(generic any) *RuntimeCreator {
	if c.svcCtx == nil {
		exception.Panicf("%w: svcCtx is nil", ErrFramework)
	}

	if generic == nil {
		exception.Panicf("%w: %w: generic is nil", ErrFramework, core.ErrArgs)
	}

	rtGeneric, ok := generic.(iRuntimeGeneric)
	if !ok {
		rtGeneric = newRuntimeInstantiation(generic)
	}

	rtGeneric.init(c.svcCtx, rtGeneric)
	c.generic = rtGeneric

	return c
}

// SetName 设置名称
func (c *RuntimeCreator) SetName(name string) *RuntimeCreator {
	c.settings.Name = name
	return c
}

// SetPersistId 设置持久化Id
func (c *RuntimeCreator) SetPersistId(id uid.Id) *RuntimeCreator {
	c.settings.PersistId = id
	return c
}

// SetPanicHandling 设置panic时的处理方式
func (c *RuntimeCreator) SetPanicHandling(autoRecover bool, reportError chan error) *RuntimeCreator {
	c.settings.AutoRecover = autoRecover
	c.settings.ReportError = reportError
	return c
}

// SetFPS 设置帧率
func (c *RuntimeCreator) SetFPS(fps float32) *RuntimeCreator {
	c.settings.FPS = fps
	return c
}

// SetProcessQueueCapacity 设置任务处理流水线大小
func (c *RuntimeCreator) SetProcessQueueCapacity(cap int) *RuntimeCreator {
	c.settings.ProcessQueueCapacity = cap
	return c
}

// SetAutoInjection 设置是否自动注入依赖的组件
func (c *RuntimeCreator) SetAutoInjection(b bool) *RuntimeCreator {
	c.settings.AutoInjection = b
	return c
}

// New 创建运行时
func (c *RuntimeCreator) New() IRuntime {
	if c.svcCtx == nil {
		exception.Panicf("%w: svcCtx is nil", ErrFramework)
	}

	generic := c.generic
	if generic == nil {
		generic = c.Setup(&RuntimeGeneric{}).generic
	}

	return reinterpret.Cast[IRuntime](runtime.Current(generic.generate(c.settings)))
}
