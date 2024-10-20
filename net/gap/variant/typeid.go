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
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework/utils/binaryutil"
	"hash/fnv"
	"io"
	"reflect"
)

// TypeId 类型Id
type TypeId uint32

// Read implements io.Reader
func (t TypeId) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUvarint(uint64(t)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write implements io.Writer
func (t *TypeId) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	v, err := bs.ReadUvarint()
	if err != nil {
		return bs.BytesRead(), err
	}
	*t = TypeId(v)

	return bs.BytesRead(), nil
}

// Size 大小
func (t TypeId) Size() int {
	return binaryutil.SizeofUvarint(uint64(t))
}

// New 创建对象指针
func (t TypeId) New() (Value, error) {
	return variantCreator.New(t)
}

// NewReflected 创建反射对象指针
func (t TypeId) NewReflected() (reflect.Value, error) {
	return variantCreator.NewReflected(t)
}

// MakeTypeId 创建类型Id
func MakeTypeId(v any) TypeId {
	hash := fnv.New32a()
	rt := reflect.ValueOf(v).Elem().Type()
	if rt.PkgPath() == "" || rt.Name() == "" {
		exception.Panic("gap-var: unsupported type")
	}
	hash.Write([]byte(types.FullNameRT(rt)))
	return TypeId(TypeId_Customize + hash.Sum32())
}

// MakeTypeIdT 创建类型Id
func MakeTypeIdT[T any]() TypeId {
	hash := fnv.New32a()
	rt := reflect.TypeFor[T]()
	if rt.PkgPath() == "" || rt.Name() == "" || !reflect.PointerTo(rt).Implements(reflect.TypeFor[Value]()) {
		exception.Panic("gap-var: unsupported type")
	}
	hash.Write([]byte(types.FullNameRT(rt)))
	return TypeId(TypeId_Customize + hash.Sum32())
}
