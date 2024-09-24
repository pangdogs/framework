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
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/core/utils/uid"
)

// CreateRuntime 创建运行时
func CreateRuntime(svcCtx service.Context) RuntimeCreator {
	if svcCtx == nil {
		panic(fmt.Errorf("%w: svcCtx is nil", core.ErrArgs))
	}
	return RuntimeCreator{
		svcCtx:  svcCtx,
		generic: nil,
		settings: _RuntimeSettings{
			Name:                 "",
			PersistId:            uid.Nil,
			AutoRecover:          svcCtx.GetAutoRecover(),
			ReportError:          svcCtx.GetReportError(),
			FPS:                  0,
			ProcessQueueCapacity: 128,
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
func (c RuntimeCreator) Setup(generic any) RuntimeCreator {
	if c.svcCtx == nil {
		panic("setting svcCtx is nil")
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

// PersistId 持久化Id
func (c RuntimeCreator) PersistId(id uid.Id) RuntimeCreator {
	c.settings.PersistId = id
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
	if c.svcCtx == nil {
		panic("setting svcCtx is nil")
	}

	generic := c.generic
	if generic == nil {
		generic = &RuntimeGenericT[RuntimeInstance]{}
	}
	generic.init(c.svcCtx, generic)

	return reinterpret.Cast[IRuntimeInstance](runtime.Current(generic.generate(c.settings)))
}
