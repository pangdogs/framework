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

// BuildRuntime 创建绑定 svcInst 的运行时构建器，并继承服务的 panic 处理配置。
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

// RuntimeCreator 保存一次运行时构建所需的装配器与选项。
// 构建器可按值复制，但不应由多个 goroutine 并发修改。
type RuntimeCreator struct {
	svcInst   IService
	assembler iRuntimeAssembler
	settings  _RuntimeSettings
}

// SetAssembler 设置运行时装配器。
// assembler 可以是自定义装配器，也可以是实现 IRuntime 的实例或 reflect.Type；
// 后两者会被转换为按需创建实例的默认装配器。
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

// SetName 设置运行时名称。
func (c *RuntimeCreator) SetName(name string) *RuntimeCreator {
	c.settings.name = name
	return c
}

// SetPersistId 设置运行时持久化 ID。
func (c *RuntimeCreator) SetPersistId(id uid.Id) *RuntimeCreator {
	c.settings.persistId = id
	return c
}

// SetMainEntity 设置主实体；主实体停用后运行时会自动终止。
func (c *RuntimeCreator) SetMainEntity(entity ec.Entity) *RuntimeCreator {
	c.settings.mainEntity = entity
	return c
}

// SetPanicHandling 设置运行时是否自动恢复 panic 以及错误报告通道。
// autoRecover 为 true 时，恢复出的错误会尝试写入 reportError。
func (c *RuntimeCreator) SetPanicHandling(autoRecover bool, reportError chan error) *RuntimeCreator {
	c.settings.autoRecover = autoRecover
	c.settings.reportError = reportError
	return c
}

// SetContinueOnActivatingEntityPanic 设置实体激活发生 panic 后是否继续运行。
// 设置为 false 时，激活失败的实体会被移除。
func (c *RuntimeCreator) SetContinueOnActivatingEntityPanic(b bool) *RuntimeCreator {
	c.settings.continueOnActivatingEntityPanic = b
	return c
}

// SetEnableFrame 设置是否启用实时帧循环。
func (c *RuntimeCreator) SetEnableFrame(b bool) *RuntimeCreator {
	c.settings.enableFrame = b
	return c
}

// SetFPS 设置启用帧循环时的目标帧率。
func (c *RuntimeCreator) SetFPS(fps float64) *RuntimeCreator {
	c.settings.fps = fps
	return c
}

// SetAutoInjection 设置实体或组件激活时是否自动注入组件依赖。
func (c *RuntimeCreator) SetAutoInjection(b bool) *RuntimeCreator {
	c.settings.autoInjection = b
	return c
}

// New 装配并自动启动运行时。
// 返回时运行时 goroutine 可能尚未完成 Starting；装配主实体失败时返回错误。
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
