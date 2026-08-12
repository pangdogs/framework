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
	"git.golaxy.org/core/ec/pt"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/iface"
	"git.golaxy.org/core/utils/meta"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/core/utils/uid"
)

// BuildEntity 创建绑定 svcInst 及 prototype 名称的实体构建器。
func BuildEntity(svcInst IService, prototype string) *EntityCreator {
	if svcInst == nil {
		exception.Panicf("%w: %w: svcInst is nil", ErrFramework, core.ErrArgs)
	}
	return &EntityCreator{
		svcInst:   svcInst,
		prototype: prototype,
	}
}

// EntityCreator 保存一次实体构建所需的运行时目标、元数据与 core 实体选项。
// 构建器可按值复制，但不应由多个 goroutine 并发修改。
type EntityCreator struct {
	svcInst   IService
	prototype string
	rtInst    IRuntime
	rtCreator *RuntimeCreator
	meta      meta.Meta
	settings  []option.Setting[ec.EntityOptions]
}

// SetRuntime 设置实体要加入的既有运行时。
// rtInst 为 nil 时清除该设置；非 nil 运行时必须属于构建器绑定的服务。
func (c *EntityCreator) SetRuntime(rtInst IRuntime) *EntityCreator {
	if c.svcInst == nil {
		exception.Panicf("%w: svcInst is nil", ErrFramework)
	}
	if rtInst != nil && rtInst.Service() != c.svcInst {
		exception.Panicf("%w: runtime service mismatch", ErrFramework)
	}
	c.rtInst = rtInst
	return c
}

// SetRuntimeCreator 设置没有指定既有运行时时使用的运行时构建器。
// nil 表示使用服务默认构建器；新实体会成为新运行时的主实体。
func (c *EntityCreator) SetRuntimeCreator(rtCreator *RuntimeCreator) *EntityCreator {
	if c.svcInst == nil {
		exception.Panicf("%w: svcInst is nil", ErrFramework)
	}
	if rtCreator != nil && rtCreator.svcInst != c.svcInst {
		exception.Panicf("%w: runtime creator service mismatch", ErrFramework)
	}
	c.rtCreator = rtCreator
	return c
}

// SetInstanceFace 设置实体实例面，用于提供自定义实体行为与反射信息。
func (c *EntityCreator) SetInstanceFace(face iface.Face[ec.Entity]) *EntityCreator {
	c.settings = append(c.settings, ec.With.InstanceFace(face))
	return c
}

// SetInstance 设置自定义实体实例。
func (c *EntityCreator) SetInstance(instance ec.Entity) *EntityCreator {
	c.settings = append(c.settings, ec.With.InstanceFace(iface.NewFaceT(instance)))
	return c
}

// SetScope 设置实体的可访问作用域。
func (c *EntityCreator) SetScope(scope ec.Scope) *EntityCreator {
	c.settings = append(c.settings, ec.With.Scope(scope))
	return c
}

// SetPersistId 设置实体持久化 ID。
func (c *EntityCreator) SetPersistId(id uid.Id) *EntityCreator {
	c.settings = append(c.settings, ec.With.PersistId(id))
	return c
}

// SetComponentAwakeOnFirstTouch 设置是否在组件首次访问时检查并调用 Awake。
func (c *EntityCreator) SetComponentAwakeOnFirstTouch(b bool) *EntityCreator {
	c.settings = append(c.settings, ec.With.ComponentAwakeOnFirstTouch(b))
	return c
}

// SetComponentUniqueID 设置是否为组件分配唯一 ID。
func (c *EntityCreator) SetComponentUniqueID(b bool) *EntityCreator {
	c.settings = append(c.settings, ec.With.ComponentUniqueID(b))
	return c
}

// SetMeta 用 dict 替换实体元数据；dict 的内容会复制到新的 Meta 中。
func (c *EntityCreator) SetMeta(dict map[string]any) *EntityCreator {
	if c.meta == nil {
		c.settings = append(c.settings, c.withMeta())
	}
	c.meta = meta.New(dict)
	return c
}

// MergeMeta 合并实体元数据，同名键会被覆盖。
func (c *EntityCreator) MergeMeta(dict map[string]any) *EntityCreator {
	for k, v := range dict {
		if c.meta == nil {
			c.settings = append(c.settings, c.withMeta())
		}
		c.meta.Add(k, v)
	}
	return c
}

// MergeMetaIfAbsent 合并实体元数据，并保留已有的同名键。
func (c *EntityCreator) MergeMetaIfAbsent(dict map[string]any) *EntityCreator {
	for k, v := range dict {
		if c.meta == nil {
			c.settings = append(c.settings, c.withMeta())
		}
		c.meta.TryAdd(k, v)
	}
	return c
}

// AssignMeta 直接绑定实体元数据；m 不会在此时复制，nil 会替换为空 Meta。
func (c *EntityCreator) AssignMeta(m meta.Meta) *EntityCreator {
	if m == nil {
		m = meta.New(nil)
	}
	if c.meta == nil {
		c.settings = append(c.settings, c.withMeta())
	}
	c.meta = m
	return c
}

// New 构造实体并将其加入指定运行时。
// 未指定运行时时会创建自动运行的新运行时，并将实体设为其主实体。
func (c *EntityCreator) New() (ec.ConcurrentEntity, error) {
	if c.svcInst == nil {
		exception.Panicf("%w: svcInst is nil", ErrFramework)
	}

	entity := pt.For(c.svcInst, c.prototype).Construct(c.settings...)

	if c.rtInst != nil {
		err := core.Submit(c.rtInst, func(rtCtx runtime.Context, _ ...any) async.Result {
			return async.NewResult(nil, rtCtx.EntityManager().AddEntity(entity))
		}).Wait(c.svcInst).Error
		if err != nil {
			return nil, err
		}
		return entity, nil
	}

	rtCreator := c.rtCreator
	if rtCreator == nil {
		rtCreator = c.svcInst.BuildRuntime()
	} else {
		rtCreator = types.Pointer(*rtCreator)
	}

	_, err := rtCreator.SetPersistId(entity.Id()).SetMainEntity(entity).New()
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// NewAsync 构造实体并返回其加入运行时的 Future。
// 使用既有运行时时加入操作由该运行时调度；新建运行时时装配过程仍在调用方同步完成。
func (c *EntityCreator) NewAsync() async.Future {
	if c.svcInst == nil {
		exception.Panicf("%w: svcInst is nil", ErrFramework)
	}

	entity := pt.For(c.svcInst, c.prototype).Construct(c.settings...)

	if c.rtInst != nil {
		return core.Submit(c.rtInst, func(rtCtx runtime.Context, _ ...any) async.Result {
			if err := rtCtx.EntityManager().AddEntity(entity); err != nil {
				return async.NewResult(nil, err)
			}
			return async.NewResult(entity, nil)
		})
	}

	rtCreator := c.rtCreator
	if rtCreator == nil {
		rtCreator = c.svcInst.BuildRuntime()
	} else {
		rtCreator = types.Pointer(*rtCreator)
	}

	_, err := rtCreator.SetPersistId(entity.Id()).SetMainEntity(entity).New()
	if err != nil {
		return async.Rejected(err)
	}

	return async.Resolved(async.NewResult(entity, nil))
}

func (c *EntityCreator) withMeta() option.Setting[ec.EntityOptions] {
	return func(o *ec.EntityOptions) {
		o.Meta = c.meta
	}
}
