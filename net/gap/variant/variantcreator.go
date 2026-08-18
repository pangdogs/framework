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

package variant

import (
	"fmt"
	"maps"
	"reflect"
	"runtime"
	"sync/atomic"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/types"
)

var (
	// ErrNotDeclared 表示指定动态值类型 ID 尚未注册。
	ErrNotDeclared = fmt.Errorf("%w: variant not declared", ErrVariant)
)

// IVariantCreator 按类型 ID 注册并构造动态值。
type IVariantCreator interface {
	// Declare 注册值指针的具体类型；重复 ID 会 panic。
	Declare(v Value)
	// New 创建指定类型 ID 对应的新值指针。
	New(typeID TypeID) (Value, error)
	// NewReflected 创建指定类型 ID 对应的新反射值。
	NewReflected(typeID TypeID) (reflect.Value, error)
}

var variantCreator = _NewVariantCreator()

// VariantCreator 返回已注册全部内置值的进程级类型构建器。
func VariantCreator() IVariantCreator {
	return variantCreator
}

func init() {
	VariantCreator().Declare(new(Int))
	VariantCreator().Declare(new(Int8))
	VariantCreator().Declare(new(Int16))
	VariantCreator().Declare(new(Int32))
	VariantCreator().Declare(new(Int64))
	VariantCreator().Declare(new(Uint))
	VariantCreator().Declare(new(Uint8))
	VariantCreator().Declare(new(Uint16))
	VariantCreator().Declare(new(Uint32))
	VariantCreator().Declare(new(Uint64))
	VariantCreator().Declare(new(Float))
	VariantCreator().Declare(new(Double))
	VariantCreator().Declare(new(Byte))
	VariantCreator().Declare(new(Bool))
	VariantCreator().Declare(new(Bytes))
	VariantCreator().Declare(new(String))
	VariantCreator().Declare(&Null{})
	VariantCreator().Declare(&Map{})
	VariantCreator().Declare(&Array{})
	VariantCreator().Declare(&Error{})
	VariantCreator().Declare(&CallChain{})
}

// _NewVariantCreator 创建空的并发安全类型构建器。
func _NewVariantCreator() IVariantCreator {
	return &_VariantCreator{}
}

// _VariantCreator 使用写时复制快照保存动态值类型映射。
type _VariantCreator struct {
	variantTypes atomic.Pointer[map[TypeID]reflect.Type]
}

// Declare 以类型 ID 注册值的元素类型。
func (c *_VariantCreator) Declare(v Value) {
	if v == nil {
		exception.Panicf("%w: %w: v is nil", ErrVariant, core.ErrArgs)
	}

	for {
		var m map[TypeID]reflect.Type

		old := c.variantTypes.Load()
		if old != nil {
			m = maps.Clone(*old)
		}

		if m == nil {
			m = make(map[TypeID]reflect.Type)
		}

		if rtype, ok := (m)[v.TypeID()]; ok {
			exception.Panicf("%w: variant type(%d) has already been declared by %q; rename the variant type or return a different TypeID", ErrVariant, v.TypeID(), types.FullNameRT(rtype))
		}

		m[v.TypeID()] = reflect.TypeOf(v).Elem()

		if c.variantTypes.CompareAndSwap(old, &m) {
			break
		}

		runtime.Gosched()
	}
}

// New 根据当前类型快照创建新的值指针。
func (c *_VariantCreator) New(typeID TypeID) (Value, error) {
	reflected, err := c.NewReflected(typeID)
	if err != nil {
		return nil, err
	}
	return reflected.Interface().(Value), nil
}

// NewReflected 根据当前类型快照创建新的反射值。
func (c *_VariantCreator) NewReflected(typeID TypeID) (reflect.Value, error) {
	m := c.variantTypes.Load()
	if m == nil || *m == nil {
		return reflect.Value{}, ErrNotDeclared
	}

	rtype, ok := (*m)[typeID]
	if !ok {
		return reflect.Value{}, ErrNotDeclared
	}

	return reflect.New(rtype), nil
}
