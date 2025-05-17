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
	"errors"
	"fmt"
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
	"reflect"
)

var (
	ErrVariant = errors.New("gap-variant") // 可变类型错误
)

// Variant 可变类型
type Variant struct {
	TypeId          TypeId           // 类型Id
	Value           Value            // 值
	Reflected       reflect.Value    // 反射值
	ReadonlyValue   ValueReader      // 只读值
	SerializedValue *SerializedValue // 已序列化值
}

// Read implements io.Reader
func (v Variant) Read(p []byte) (int, error) {
	if !v.Valid() {
		return 0, fmt.Errorf("%w: invalid variant", ErrVariant)
	}

	bs := binaryutil.NewBigEndianStream(p)

	if _, err := binaryutil.CopyToByteStream(&bs, v.TypeId); err != nil {
		return bs.BytesWritten(), err
	}

	if v.Readonly() {
		if v.Serialized() {
			if _, err := binaryutil.CopyToByteStream(&bs, v.SerializedValue); err != nil {
				return bs.BytesWritten(), err
			}
		} else {
			if _, err := binaryutil.CopyToByteStream(&bs, v.ReadonlyValue); err != nil {
				return bs.BytesWritten(), err
			}
		}
	} else {
		if _, err := binaryutil.CopyToByteStream(&bs, v.Value); err != nil {
			return bs.BytesWritten(), err
		}
	}

	return bs.BytesWritten(), io.EOF
}

// Write implements io.Writer
func (v *Variant) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if _, err := bs.WriteTo(&v.TypeId); err != nil {
		return bs.BytesRead(), err
	}

	reflected, err := v.TypeId.NewReflected()
	if err != nil {
		return bs.BytesRead(), err
	}

	value := reflected.Interface().(Value)
	if _, err := bs.WriteTo(value); err != nil {
		return bs.BytesRead(), err
	}

	v.ReadonlyValue = nil
	v.SerializedValue = nil
	v.Value = value
	v.Reflected = reflected

	return bs.BytesRead(), nil
}

// Size 大小
func (v Variant) Size() int {
	if !v.Valid() {
		return 0
	}

	n := v.TypeId.Size()

	if v.Readonly() {
		if v.Serialized() {
			n += v.SerializedValue.Size()
		} else {
			n += v.ReadonlyValue.Size()
		}
	} else {
		if v.Value != nil {
			n += v.Value.Size()
		}
	}

	return n
}

// Readonly 是否只读
func (v Variant) Readonly() bool {
	return v.Serialized() || v.ReadonlyValue != nil
}

// Serialized 是否已序列化
func (v Variant) Serialized() bool {
	return v.SerializedValue != nil
}

// Valid 是否有效
func (v Variant) Valid() bool {
	if v.Readonly() {
		if v.Serialized() {
			return v.TypeId == v.SerializedValue.TypeId()
		} else {
			return v.TypeId == v.ReadonlyValue.TypeId()
		}
	}
	if v.Value != nil {
		return v.TypeId == v.Value.TypeId()
	}
	return false
}
