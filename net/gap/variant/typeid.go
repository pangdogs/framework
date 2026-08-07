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
	"hash/fnv"
	"io"
	"reflect"

	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework/utils/binaryutil"
)

// TypeId 标识 GAP 动态值的具体类型。
type TypeId uint32

// Read 将类型 ID 编码到 p。
func (t TypeId) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint32(uint32(t)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码类型 ID。
func (t *TypeId) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	v, err := bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}
	*t = TypeId(v)

	return bs.BytesRead(), nil
}

// Size 返回类型 ID 的固定编码字节数。
func (t TypeId) Size() int {
	return binaryutil.SizeofUint32
}

// New 使用进程级类型构建器创建此类型的新值。
func (t TypeId) New() (Value, error) {
	return variantCreator.New(t)
}

// NewReflected 使用进程级类型构建器创建此类型的新反射值。
func (t TypeId) NewReflected() (reflect.Value, error) {
	return variantCreator.NewReflected(t)
}

// GenTypeId 根据具名 Value 类型的完整名称生成自定义类型 ID。
func GenTypeId(v any) TypeId {
	hash := fnv.New32a()
	rt := reflect.ValueOf(v).Elem().Type()
	if rt.PkgPath() == "" || rt.Name() == "" {
		exception.Panicf("%w: unsupported type", ErrVariant)
	}
	hash.Write([]byte(types.FullNameRT(rt)))
	return TypeId(TypeId_Customize + hash.Sum32())
}

// GenTypeIdT 根据具名类型 T 的完整名称生成自定义类型 ID；*T 必须实现 Value。
func GenTypeIdT[T any]() TypeId {
	hash := fnv.New32a()
	rt := reflect.TypeFor[T]()
	if rt.PkgPath() == "" || rt.Name() == "" || !reflect.PointerTo(rt).Implements(reflect.TypeFor[Value]()) {
		exception.Panicf("%w: unsupported type", ErrVariant)
	}
	hash.Write([]byte(types.FullNameRT(rt)))
	return TypeId(TypeId_Customize + hash.Sum32())
}
