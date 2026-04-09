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

// BuildEntity 创建实体
func BuildEntity(svcInst IService, prototype string) *EntityCreator {
	if svcInst == nil {
		exception.Panicf("%w: %w: svcInst is nil", ErrFramework, core.ErrArgs)
	}
	return &EntityCreator{
		svcInst:   svcInst,
		prototype: prototype,
	}
}

// EntityCreator 实体构建器
type EntityCreator struct {
	svcInst   IService
	prototype string
	rtInst    IRuntime
	rtCreator *RuntimeCreator
	meta      meta.Meta
	settings  []option.Setting[ec.EntityOptions]
}

// SetRuntime 设置运行时（值不为nil时，创建的新实体将会加入此处设置的运行时，不会创建新运行时）
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

// SetRuntimeCreator 设置运行时构建器（如果未设置运行时，将会使用此处设置的构建器创建新运行时，此处值为nil时，将会使用默认运行时构建器；创建出的新实体将会自动成为新运行时的主实体）
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

// SetInstanceFace 设置实例，用于扩展实体能力
func (c *EntityCreator) SetInstanceFace(face iface.Face[ec.Entity]) *EntityCreator {
	c.settings = append(c.settings, ec.With.InstanceFace(face))
	return c
}

// SetInstance 设置实例，用于扩展实体能力
func (c *EntityCreator) SetInstance(instance ec.Entity) *EntityCreator {
	c.settings = append(c.settings, ec.With.InstanceFace(iface.NewFaceT(instance)))
	return c
}

// SetScope 设置实体的可访问作用域
func (c *EntityCreator) SetScope(scope ec.Scope) *EntityCreator {
	c.settings = append(c.settings, ec.With.Scope(scope))
	return c
}

// SetPersistId 设置实体持久化Id
func (c *EntityCreator) SetPersistId(id uid.Id) *EntityCreator {
	c.settings = append(c.settings, ec.With.PersistId(id))
	return c
}

// SetComponentAwakeOnFirstTouch 设置开启组件被首次访问时，检测并调用Awake()
func (c *EntityCreator) SetComponentAwakeOnFirstTouch(b bool) *EntityCreator {
	c.settings = append(c.settings, ec.With.ComponentAwakeOnFirstTouch(b))
	return c
}

// SetComponentUniqueID 设置开启组件唯一Id
func (c *EntityCreator) SetComponentUniqueID(b bool) *EntityCreator {
	c.settings = append(c.settings, ec.With.ComponentUniqueID(b))
	return c
}

// SetMeta 设置Meta信息
func (c *EntityCreator) SetMeta(dict map[string]any) *EntityCreator {
	if c.meta == nil {
		c.settings = append(c.settings, c.withMeta())
	}
	c.meta = meta.New(dict)
	return c
}

// MergeMeta 合并Meta信息，如果存在则覆盖
func (c *EntityCreator) MergeMeta(dict map[string]any) *EntityCreator {
	for k, v := range dict {
		if c.meta == nil {
			c.settings = append(c.settings, c.withMeta())
		}
		c.meta.Add(k, v)
	}
	return c
}

// MergeMetaIfAbsent 合并Meta信息，如果存在则跳过
func (c *EntityCreator) MergeMetaIfAbsent(dict map[string]any) *EntityCreator {
	for k, v := range dict {
		if c.meta == nil {
			c.settings = append(c.settings, c.withMeta())
		}
		c.meta.TryAdd(k, v)
	}
	return c
}

// AssignMeta 赋值Meta信息
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

// New 创建实体
func (c *EntityCreator) New() (ec.ConcurrentEntity, error) {
	if c.svcInst == nil {
		exception.Panicf("%w: svcInst is nil", ErrFramework)
	}

	entity := pt.For(c.svcInst, c.prototype).Construct(c.settings...)

	if c.rtInst != nil {
		err := core.CallAsync(c.rtInst, func(rtCtx runtime.Context, _ ...any) async.Result {
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

// NewAsync 创建实体
func (c *EntityCreator) NewAsync() async.Future {
	if c.svcInst == nil {
		exception.Panicf("%w: svcInst is nil", ErrFramework)
	}

	entity := pt.For(c.svcInst, c.prototype).Construct(c.settings...)

	if c.rtInst != nil {
		return core.CallAsync(c.rtInst, func(rtCtx runtime.Context, _ ...any) async.Result {
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

	resultFuture := async.NewFutureChan()

	_, err := rtCreator.SetPersistId(entity.Id()).SetMainEntity(entity).New()
	if err != nil {
		return async.Return(resultFuture, async.NewResult(nil, err))
	}

	return async.Return(resultFuture, async.NewResult(entity, nil))
}

func (c *EntityCreator) withMeta() option.Setting[ec.EntityOptions] {
	return func(o *ec.EntityOptions) {
		o.Meta = c.meta
	}
}
