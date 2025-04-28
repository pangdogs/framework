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
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/iface"
	"git.golaxy.org/core/utils/meta"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
)

// BuildEntityAsync 创建实体
func BuildEntityAsync(svcCtx service.Context, prototype string) *EntityCreatorAsync {
	if svcCtx == nil {
		exception.Panicf("%w: %w: svcCtx is nil", ErrFramework, core.ErrArgs)
	}
	return &EntityCreatorAsync{
		ctx:       svcCtx,
		prototype: prototype,
	}
}

// EntityCreatorAsync 实体构建器
type EntityCreatorAsync struct {
	ctx       service.Context
	prototype string
	rtInst    IRuntime
	rtCreator *RuntimeCreator
	parentId  uid.Id
	meta      meta.Meta
	settings  []option.Setting[ec.EntityOptions]
}

// SetRuntime 设置运行时（优先使用）
func (c *EntityCreatorAsync) SetRuntime(rtInst IRuntime) *EntityCreatorAsync {
	c.rtInst = rtInst
	return c
}

// SetRuntimeCreator 设置运行时构建器
func (c *EntityCreatorAsync) SetRuntimeCreator(rtCreator *RuntimeCreator) *EntityCreatorAsync {
	c.rtCreator = rtCreator
	return c
}

// SetInstanceFace 设置实例，用于扩展实体能力
func (c *EntityCreatorAsync) SetInstanceFace(face iface.Face[ec.Entity]) *EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.InstanceFace(face))
	return c
}

// SetInstance 设置实例，用于扩展实体能力
func (c *EntityCreatorAsync) SetInstance(instance ec.Entity) *EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.InstanceFace(iface.MakeFaceT(instance)))
	return c
}

// SetScope 设置实体的可访问作用域
func (c *EntityCreatorAsync) SetScope(scope ec.Scope) *EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.Scope(scope))
	return c
}

// SetPersistId 设置实体持久化Id
func (c *EntityCreatorAsync) SetPersistId(id uid.Id) *EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.PersistId(id))
	return c
}

// SetComponentNameIndexing 设置是否开启组件名称索引
func (c *EntityCreatorAsync) SetComponentNameIndexing(b bool) *EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.ComponentNameIndexing(b))
	return c
}

// SetComponentAwakeOnFirstTouch 设置开启组件被首次访问时，检测并调用Awake()
func (c *EntityCreatorAsync) SetComponentAwakeOnFirstTouch(b bool) *EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.ComponentAwakeOnFirstTouch(b))
	return c
}

// SetComponentUniqueID 设置开启组件唯一Id
func (c *EntityCreatorAsync) SetComponentUniqueID(b bool) *EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.ComponentUniqueID(b))
	return c
}

// SetMeta 设置Meta信息
func (c *EntityCreatorAsync) SetMeta(dict map[string]any) *EntityCreatorAsync {
	if c.meta == nil {
		c.settings = append(c.settings, ec.With.Meta(c.meta))
	}
	c.meta = meta.M(dict)
	return c
}

// MergeMeta 合并Meta信息，如果存在则覆盖
func (c *EntityCreatorAsync) MergeMeta(dict map[string]any) *EntityCreatorAsync {
	for k, v := range dict {
		if c.meta == nil {
			c.settings = append(c.settings, ec.With.Meta(c.meta))
		}
		c.meta.Add(k, v)
	}
	return c
}

// MergeMetaIfAbsent 合并Meta信息，如果存在则跳过
func (c *EntityCreatorAsync) MergeMetaIfAbsent(dict map[string]any) *EntityCreatorAsync {
	for k, v := range dict {
		if c.meta == nil {
			c.settings = append(c.settings, ec.With.Meta(c.meta))
		}
		c.meta.TryAdd(k, v)
	}
	return c
}

// AssignMeta 赋值Meta信息
func (c *EntityCreatorAsync) AssignMeta(m meta.Meta) *EntityCreatorAsync {
	if m == nil {
		m = meta.M(nil)
	}
	if c.meta == nil {
		c.settings = append(c.settings, ec.With.Meta(c.meta))
	}
	c.meta = m
	return c
}

// SetParentId 设置父实体Id
func (c *EntityCreatorAsync) SetParentId(id uid.Id) *EntityCreatorAsync {
	c.parentId = id
	return c
}

// New 创建实体
func (c *EntityCreatorAsync) New() (ec.ConcurrentEntity, error) {
	if c.ctx == nil {
		exception.Panicf("%w: ctx is nil", ErrFramework)
	}

	entity := pt.For(c.ctx, c.prototype).Construct(c.settings...)

	rtInst := c.rtInst
	if rtInst == nil {
		rtCreator := c.rtCreator
		if rtCreator == nil {
			rtCreator = BuildRuntime(c.ctx)
		}
		rtInst = rtCreator.SetPersistId(entity.GetId()).New()
	}

	err := core.CallAsync(rtInst, func(rtCtx runtime.Context, _ ...any) async.Ret {
		if c.parentId.IsNil() {
			if err := rtCtx.GetEntityManager().AddEntity(entity); err != nil {
				return async.MakeRet(nil, err)
			}
		} else {
			if err := rtCtx.GetEntityTree().AddNode(entity, c.parentId); err != nil {
				return async.MakeRet(nil, err)
			}
		}
		return async.VoidRet
	}).Wait(c.ctx).Error
	if err != nil {
		if c.rtInst == nil {
			rtInst.Terminate()
		}
		return nil, err
	}

	return entity, nil
}

// NewAsync 创建实体
func (c *EntityCreatorAsync) NewAsync() async.AsyncRetT[ec.ConcurrentEntity] {
	if c.ctx == nil {
		exception.Panicf("%w: ctx is nil", ErrFramework)
	}

	entity := pt.For(c.ctx, c.prototype).Construct(c.settings...)

	rtInst := c.rtInst
	if rtInst == nil {
		rtCreator := c.rtCreator
		if rtCreator == nil {
			rtCreator = BuildRuntime(c.ctx)
		}
		rtInst = rtCreator.SetPersistId(entity.GetId()).New()
	}

	asyncRet := core.CallAsync(rtInst, func(rtCtx runtime.Context, _ ...any) async.Ret {
		if c.parentId.IsNil() {
			if err := rtCtx.GetEntityManager().AddEntity(entity); err != nil {
				return async.MakeRet(nil, err)
			}
		} else {
			if err := rtCtx.GetEntityTree().AddNode(entity, c.parentId); err != nil {
				return async.MakeRet(nil, err)
			}
		}
		return async.VoidRet
	})

	asyncRetT := async.MakeAsyncRetT[ec.ConcurrentEntity]()
	go func() {
		ret := asyncRet.Wait(c.ctx)
		if ret.OK() {
			async.ReturnT(asyncRetT, async.MakeRetT[ec.ConcurrentEntity](entity, nil))
		} else {
			if c.rtInst == nil {
				rtInst.Terminate()
			}
			async.ReturnT(asyncRetT, async.MakeRetT[ec.ConcurrentEntity](nil, ret.Error))
		}
	}()

	return asyncRetT
}
