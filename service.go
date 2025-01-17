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

// IServiceInstantiation 服务实例化接口
type IServiceInstantiation interface {
	Instantiation() IServiceInstance
}

// NewServiceInstantiation 创建服务类型实例化
func NewServiceInstantiation(svcInst any) *ServiceInstantiation {
	if svcInst == nil {
		exception.Panicf("%w: %w: svcInst is nil", ErrFramework)
	}

	svcInstRT, ok := svcInst.(reflect.Type)
	if !ok {
		svcInstRT = reflect.ValueOf(svcInst).Type()
	}

	for svcInstRT.Kind() == reflect.Pointer {
		svcInstRT = svcInstRT.Elem()
	}

	if svcInstRT.PkgPath() == "" || svcInstRT.Name() == "" || !reflect.PointerTo(svcInstRT).Implements(reflect.TypeFor[IServiceInstance]()) {
		exception.Panicf("%w: unsupported type", ErrFramework)
	}

	return &ServiceInstantiation{
		serviceInstanceRT: svcInstRT,
	}
}

// NewServiceInstantiationT 创建服务类型实例化
func NewServiceInstantiationT[T any]() *ServiceInstantiation {
	return NewServiceInstantiation(types.ZeroT[T]())
}

// ServiceInstantiation 服务类型实例化
type ServiceInstantiation struct {
	ServiceGeneric
	serviceInstanceRT reflect.Type
}

func (s *ServiceInstantiation) Instantiation() IServiceInstance {
	if s.serviceInstanceRT == nil {
		return &ServiceInstance{}
	}
	return reflect.New(s.serviceInstanceRT).Interface().(IServiceInstance)
}
