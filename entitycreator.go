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
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/pt"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/iface"
	"git.golaxy.org/core/utils/meta"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
)

// CreateConcurrentEntity 创建实体
func CreateConcurrentEntity(svcCtx service.Context, prototype string) ConcurrentEntityCreator {
	if svcCtx == nil {
		panic(fmt.Errorf("%w: svcCtx is nil", core.ErrArgs))
	}
	return ConcurrentEntityCreator{
		ctx:       svcCtx,
		prototype: prototype,
	}
}

// ConcurrentEntityCreator 实体构建器
type ConcurrentEntityCreator struct {
	ctx       service.Context
	prototype string
	rtInst    IRuntimeInstance
	rtCreator RuntimeCreator
	parentId  uid.Id
	settings  []option.Setting[ec.EntityOptions]
}

// Runtime 设置运行时（优先使用）
func (c ConcurrentEntityCreator) Runtime(rtInst IRuntimeInstance) ConcurrentEntityCreator {
	c.rtInst = rtInst
	return c
}

// RuntimeCreator 设置运行时构建器
func (c ConcurrentEntityCreator) RuntimeCreator(rtCreator RuntimeCreator) ConcurrentEntityCreator {
	c.rtCreator = rtCreator
	return c
}

// InstanceFace 实例，用于扩展实体能力
func (c ConcurrentEntityCreator) InstanceFace(face iface.Face[ec.Entity]) ConcurrentEntityCreator {
	c.settings = append(c.settings, ec.With.InstanceFace(face))
	return c
}

// Instance 实例，用于扩展实体能力
func (c ConcurrentEntityCreator) Instance(instance ec.Entity) ConcurrentEntityCreator {
	c.settings = append(c.settings, ec.With.InstanceFace(iface.MakeFaceT(instance)))
	return c
}

// Scope 设置实体的可访问作用域
func (c ConcurrentEntityCreator) Scope(scope ec.Scope) ConcurrentEntityCreator {
	c.settings = append(c.settings, ec.With.Scope(scope))
	return c
}

// PersistId 设置实体持久化Id
func (c ConcurrentEntityCreator) PersistId(id uid.Id) ConcurrentEntityCreator {
	c.settings = append(c.settings, ec.With.PersistId(id))
	return c
}

// AwakeOnFirstAccess 设置开启组件被首次访问时，检测并调用Awake()
func (c ConcurrentEntityCreator) AwakeOnFirstAccess(b bool) ConcurrentEntityCreator {
	c.settings = append(c.settings, ec.With.AwakeOnFirstAccess(b))
	return c
}

// Meta 设置Meta信息
func (c ConcurrentEntityCreator) Meta(m meta.Meta) ConcurrentEntityCreator {
	c.settings = append(c.settings, ec.With.Meta(m))
	return c
}

// ParentId 设置父实体Id
func (c ConcurrentEntityCreator) ParentId(id uid.Id) ConcurrentEntityCreator {
	c.parentId = id
	return c
}

// Spawn 创建实体
func (c ConcurrentEntityCreator) Spawn() (ec.ConcurrentEntity, error) {
	if c.ctx == nil {
		panic(" setting ctx is nil")
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
			if err := rtCtx.GetEntityMgr().AddEntity(entity); err != nil {
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
func (c ConcurrentEntityCreator) SpawnAsync() async.AsyncRetT[ec.ConcurrentEntity] {
	if c.ctx == nil {
		panic(" setting ctx is nil")
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
			if err := rtCtx.GetEntityMgr().AddEntity(entity); err != nil {
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
