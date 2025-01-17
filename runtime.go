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
	"git.golaxy.org/core/utils/exception"
	"reflect"
)

// IRuntimeInstantiation 运行时实例化接口
type IRuntimeInstantiation interface {
	Instantiation() IRuntimeInstance
}

func newRuntimeInstantiation(rtInst any) *_RuntimeInstantiation {
	if rtInst == nil {
		exception.Panicf("%w: %w: rtInst is nil", ErrFramework, core.ErrArgs)
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

	return &_RuntimeInstantiation{
		runtimeInstanceRT: rtInstRT,
	}
}
 
type _RuntimeInstantiation struct {
	RuntimeGeneric
	runtimeInstanceRT reflect.Type
}

func (r *_RuntimeInstantiation) Instantiation() IRuntimeInstance {
	if r.runtimeInstanceRT == nil {
		exception.Panicf("%w: runtimeInstanceRT is nil", ErrFramework)
	}
	return reflect.New(r.runtimeInstanceRT).Interface().(IRuntimeInstance)
}
