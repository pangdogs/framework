package framework

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/pt"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/iface"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/core/util/uid"
)

// CreateConcurrentEntity 创建实体
func CreateConcurrentEntity(ctx service.Context, prototype string) ConcurrentEntityCreator {
	if ctx == nil {
		panic(fmt.Errorf("%w: ctx is nil", core.ErrArgs))
	}
	return ConcurrentEntityCreator{
		ctx:       ctx,
		prototype: prototype,
	}
}

// ConcurrentEntityCreator 实体构建器
type ConcurrentEntityCreator struct {
	ctx       service.Context
	prototype string
	rt        core.Runtime
	rtCreator RuntimeCreator
	parentId  uid.Id
	settings  []option.Setting[ec.EntityOptions]
}

// Runtime 设置运行时（优先使用）
func (c ConcurrentEntityCreator) Runtime(rt core.Runtime) ConcurrentEntityCreator {
	c.rt = rt
	return c
}

// RuntimeCreator 设置运行时构建器
func (c ConcurrentEntityCreator) RuntimeCreator(rtCreator RuntimeCreator) ConcurrentEntityCreator {
	c.rtCreator = rtCreator
	return c
}

// CompositeFace 设置扩展者，在扩展实体自身能力时使用
func (c ConcurrentEntityCreator) CompositeFace(face iface.Face[ec.Entity]) ConcurrentEntityCreator {
	c.settings = append(c.settings, ec.With.CompositeFace(face))
	return c
}

// Composite 设置扩展者，在扩展实体自身能力时使用
func (c ConcurrentEntityCreator) Composite(composite ec.Entity) ConcurrentEntityCreator {
	c.settings = append(c.settings, ec.With.CompositeFace(iface.MakeFace(composite)))
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
func (c ConcurrentEntityCreator) Meta(m ec.Meta) ConcurrentEntityCreator {
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

	rt := c.rt
	if rt == nil {
		if c.rtCreator.servCtx != nil {
			rt = c.rtCreator.Spawn()
		} else {
			rt = CreateRuntime(c.ctx).Spawn()
		}
	}

	err := core.Async(rt, func(ctx runtime.Context, _ ...any) runtime.Ret {
		if c.parentId.IsNil() {
			if err := ctx.GetEntityMgr().AddEntity(entity); err != nil {
				return runtime.MakeRet(nil, err)
			}
		} else {
			if err := ctx.GetEntityTree().AddNode(entity, c.parentId); err != nil {
				return runtime.MakeRet(nil, err)
			}
		}
		return runtime.VoidRet
	}).Wait(c.ctx).Error
	if err != nil {
		if c.rt == nil {
			rt.Terminate()
		}
		return nil, err
	}

	return entity, nil
}
