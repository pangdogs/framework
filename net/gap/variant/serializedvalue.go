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
	"git.golaxy.org/core"
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
)

// MakeSerializedValue 创建已序列化值
func MakeSerializedValue(v ValueReader) (ret *SerializedValue, err error) {
	if v == nil {
		return nil, fmt.Errorf("%w: %w: v is nil", ErrVariant, core.ErrArgs)
	}

	sv := &SerializedValue{
		Type: v.TypeId(),
	}

	size := v.Size()
	if size > 0 {
		buf := binaryutil.MakeRecycleBytes(size)
		defer func() {
			if ret == nil {
				buf.Release()
			}
		}()

		if _, err := binaryutil.CopyToBuff(buf.Data(), v); err != nil {
			return nil, err
		}

		sv.Data = buf

	} else {
		sv.Data = binaryutil.NilRecycleBytes
	}

	return sv, nil
}

// SerializedValue 已序列化值
type SerializedValue struct {
	Type TypeId                  // 类型Id
	Data binaryutil.RecycleBytes // 数据
}

// Read implements io.Reader
func (v *SerializedValue) Read(p []byte) (int, error) {
	if len(p) < len(v.Data.Data()) {
		return 0, io.ErrShortWrite
	}
	return copy(p, v.Data.Data()), io.EOF
}

// Size 大小
func (v *SerializedValue) Size() int {
	return len(v.Data.Data())
}

// TypeId 类型
func (v *SerializedValue) TypeId() TypeId {
	return v.Type
}

// Indirect 原始值
func (v *SerializedValue) Indirect() any {
	return v
}

// Release 释放资源
func (v *SerializedValue) Release() {
	v.Data.Release()
}
