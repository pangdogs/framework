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
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/core/utils/uid"
)

// BuildRuntime 创建运行时
func BuildRuntime(svcInst IService) *RuntimeCreator {
	if svcInst == nil {
		exception.Panicf("%w: %w: svcInst is nil", ErrFramework, core.ErrArgs)
	}
	return &RuntimeCreator{
		svcInst:   svcInst,
		assembler: svcInst.(iService).getRuntimeAssembler(),
		settings: _RuntimeSettings{
			name:                            "",
			persistId:                       uid.Nil,
			mainEntity:                      nil,
			autoRecover:                     svcInst.AutoRecover(),
			reportError:                     svcInst.ReportError(),
			continueOnActivatingEntityPanic: false,
			enableFrame:                     false,
			fps:                             30,
			autoInjection:                   true,
		},
	}
}

// RuntimeCreator 运行时实例构建器
type RuntimeCreator struct {
	svcInst   IService
	assembler iRuntimeAssembler
	settings  _RuntimeSettings
}

// SetAssembler 设置运行时实例装配器
func (c *RuntimeCreator) SetAssembler(assembler any) *RuntimeCreator {
	if c.svcInst == nil {
		exception.Panicf("%w: svcInst is nil", ErrFramework)
	}

	if assembler == nil {
		exception.Panicf("%w: %w: assembler is nil", ErrFramework, core.ErrArgs)
	}

	assemblerInst, ok := assembler.(iRuntimeAssembler)
	if !ok {
		assemblerInst = newRuntimeInstantiator(assembler)
	}
	assemblerInst.init(c.svcInst, assemblerInst)

	c.assembler = assemblerInst

	return c
}

// SetName 设置名称
func (c *RuntimeCreator) SetName(name string) *RuntimeCreator {
	c.settings.name = name
	return c
}

// SetPersistId 设置持久化Id
func (c *RuntimeCreator) SetPersistId(id uid.Id) *RuntimeCreator {
	c.settings.persistId = id
	return c
}

// SetMainEntity 设置主实体（主实体和运行时生命周期绑定，主实体销毁时，运行时将会停止运行）
func (c *RuntimeCreator) SetMainEntity(entity ec.Entity) *RuntimeCreator {
	c.settings.mainEntity = entity
	return c
}

// SetPanicHandling 设置panic时的处理方式
func (c *RuntimeCreator) SetPanicHandling(autoRecover bool, reportError chan error) *RuntimeCreator {
	c.settings.autoRecover = autoRecover
	c.settings.reportError = reportError
	return c
}

// SetContinueOnActivatingEntityPanic 设置激活实体时发生panic是否继续，不继续将会主动删除实体
func (c *RuntimeCreator) SetContinueOnActivatingEntityPanic(b bool) *RuntimeCreator {
	c.settings.continueOnActivatingEntityPanic = b
	return c
}

// SetEnableFrame 设置是否启用帧循环
func (c *RuntimeCreator) SetEnableFrame(b bool) *RuntimeCreator {
	c.settings.enableFrame = b
	return c
}

// SetFPS 设置帧率
func (c *RuntimeCreator) SetFPS(fps float64) *RuntimeCreator {
	c.settings.fps = fps
	return c
}

// SetAutoInjection 设置是否自动注入依赖的组件
func (c *RuntimeCreator) SetAutoInjection(b bool) *RuntimeCreator {
	c.settings.autoInjection = b
	return c
}

// New 创建运行时
func (c *RuntimeCreator) New() (IRuntime, error) {
	if c.svcInst == nil {
		exception.Panicf("%w: svcInst is nil", ErrFramework)
	}
	if c.assembler == nil {
		exception.Panicf("%w: assembler is nil", ErrFramework)
	}
	rt, err := c.assembler.assemble(c.settings)
	if err != nil {
		return nil, err
	}
	return reinterpret.Cast[IRuntime](runtime.Current(rt)), nil
}
