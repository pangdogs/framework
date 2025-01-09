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

// CreateEntityAsync 创建实体
func CreateEntityAsync(svcCtx service.Context, prototype string) EntityCreatorAsync {
	if svcCtx == nil {
		exception.Panicf("%w: %w: svcCtx is nil", ErrFramework, core.ErrArgs)
	}
	return EntityCreatorAsync{
		ctx:       svcCtx,
		prototype: prototype,
	}
}

// EntityCreatorAsync 实体构建器
type EntityCreatorAsync struct {
	ctx       service.Context
	prototype string
	rtInst    IRuntimeInstance
	rtCreator RuntimeCreator
	parentId  uid.Id
	settings  []option.Setting[ec.EntityOptions]
}

// Runtime 设置运行时（优先使用）
func (c EntityCreatorAsync) Runtime(rtInst IRuntimeInstance) EntityCreatorAsync {
	c.rtInst = rtInst
	return c
}

// RuntimeCreator 设置运行时构建器
func (c EntityCreatorAsync) RuntimeCreator(rtCreator RuntimeCreator) EntityCreatorAsync {
	c.rtCreator = rtCreator
	return c
}

// InstanceFace 实例，用于扩展实体能力
func (c EntityCreatorAsync) InstanceFace(face iface.Face[ec.Entity]) EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.InstanceFace(face))
	return c
}

// Instance 实例，用于扩展实体能力
func (c EntityCreatorAsync) Instance(instance ec.Entity) EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.InstanceFace(iface.MakeFaceT(instance)))
	return c
}

// Scope 设置实体的可访问作用域
func (c EntityCreatorAsync) Scope(scope ec.Scope) EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.Scope(scope))
	return c
}

// PersistId 设置实体持久化Id
func (c EntityCreatorAsync) PersistId(id uid.Id) EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.PersistId(id))
	return c
}

// ComponentNameIndexing 是否开启组件名称索引
func (c EntityCreatorAsync) ComponentNameIndexing(b bool) EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.ComponentNameIndexing(b))
	return c
}

// ComponentAwakeOnFirstTouch 设置开启组件被首次访问时，检测并调用Awake()
func (c EntityCreatorAsync) ComponentAwakeOnFirstTouch(b bool) EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.ComponentAwakeOnFirstTouch(b))
	return c
}

// ComponentUniqueID 设置开启组件唯一Id
func (c EntityCreatorAsync) ComponentUniqueID(b bool) EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.ComponentUniqueID(b))
	return c
}

// Meta 设置Meta信息
func (c EntityCreatorAsync) Meta(m meta.Meta) EntityCreatorAsync {
	c.settings = append(c.settings, ec.With.Meta(m))
	return c
}

// ParentId 设置父实体Id
func (c EntityCreatorAsync) ParentId(id uid.Id) EntityCreatorAsync {
	c.parentId = id
	return c
}

// Spawn 创建实体
func (c EntityCreatorAsync) Spawn() (ec.ConcurrentEntity, error) {
	if c.ctx == nil {
		exception.Panicf("%w: ctx is nil", ErrFramework)
	}

	entity := pt.For(c.ctx, c.prototype).Construct(c.settings...)

	rtInst := c.rtInst
	if rtInst == nil {
		if c.rtCreator.svcCtx != nil {
			rtInst = c.rtCreator.Spawn()
		} else {
			rtInst = CreateRuntime(c.ctx).PersistId(entity.GetId()).Spawn()
		}
	}

	err := core.Async(rtInst, func(rtCtx runtime.Context, _ ...any) async.Ret {
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

// SpawnAsync 创建实体
func (c EntityCreatorAsync) SpawnAsync() async.AsyncRetT[ec.ConcurrentEntity] {
	if c.ctx == nil {
		exception.Panicf("%w: ctx is nil", ErrFramework)
	}

	entity := pt.For(c.ctx, c.prototype).Construct(c.settings...)

	rtInst := c.rtInst
	if rtInst == nil {
		if c.rtCreator.svcCtx != nil {
			rtInst = c.rtCreator.Spawn()
		} else {
			rtInst = CreateRuntime(c.ctx).PersistId(entity.GetId()).Spawn()
		}
	}

	asyncRet := core.Async(rtInst, func(rtCtx runtime.Context, _ ...any) async.Ret {
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

	ch := async.MakeAsyncRetT[ec.ConcurrentEntity]()
	go func() {
		ret := asyncRet.Wait(c.ctx)
		if ret.OK() {
			ch <- async.MakeRetT[ec.ConcurrentEntity](entity, nil)
		} else {
			if c.rtInst == nil {
				rtInst.Terminate()
			}
			ch <- async.MakeRetT[ec.ConcurrentEntity](nil, ret.Error)
		}
		close(ch)
	}()
	return ch
}
