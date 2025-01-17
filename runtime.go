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
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/types"
	"reflect"
)

// IRuntimeInstantiation 运行时实例化接口
type IRuntimeInstantiation interface {
	Instantiation() IRuntimeInstance
}

// NewRuntimeInstantiation 创建运行时类型实例化
func NewRuntimeInstantiation(rtInst any) *RuntimeInstantiation {
	if rtInst == nil {
		exception.Panicf("%w: %w: rtInst is nil", ErrFramework)
	}

	rtInstRT, ok := rtInst.(reflect.Type)
	if !ok {
		rtInstRT = reflect.ValueOf(rtInst).Type()
	}

	for rtInstRT.Kind() == reflect.Pointer {
		rtInstRT = rtInstRT.Elem()
	}

	if rtInstRT.PkgPath() == "" || rtInstRT.Name() == "" || !reflect.PointerTo(rtInstRT).Implements(reflect.TypeFor[IRuntimeInstance]()) {
		exception.Panicf("%w: unsupported type", ErrFramework)
	}

	return &RuntimeInstantiation{
		runtimeInstanceRT: rtInstRT,
	}
}

// NewRuntimeInstantiationT 创建运行时类型实例化
func NewRuntimeInstantiationT[T any]() *RuntimeInstantiation {
	return NewRuntimeInstantiation(types.ZeroT[T]())
}

// RuntimeInstantiation 运行时类型实例化
type RuntimeInstantiation struct {
	RuntimeGeneric
	runtimeInstanceRT reflect.Type
}

func (r *RuntimeInstantiation) Instantiation() IRuntimeInstance {
	if r.runtimeInstanceRT == nil {
		return &RuntimeInstance{}
	}
	return reflect.New(r.runtimeInstanceRT).Interface().(IRuntimeInstance)
}
