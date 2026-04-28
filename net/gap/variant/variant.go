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
	"io"
	"reflect"

	"git.golaxy.org/core"
	"git.golaxy.org/framework/utils/binaryutil"
)

var (
	ErrVariant = errors.New("gap-variant") // 变体错误
)

// NewVariant 创建变体
func NewVariant(v ReadableValue) (Variant, error) {
	if v == nil {
		return Variant{}, fmt.Errorf("%w: %w: v is nil", ErrVariant, core.ErrArgs)
	}
	return Variant{
		TypeId: v.TypeId(),
		Value:  v,
	}, nil
}

// Variant 变体
type Variant struct {
	TypeId    TypeId        // 类型Id
	Value     ReadableValue // 值
	Reflected reflect.Value // 反射值
}

// Read implements io.Reader
func (v Variant) Read(p []byte) (int, error) {
	if !v.IsValid() {
		return 0, fmt.Errorf("%w: invalid variant", ErrVariant)
	}

	bs := binaryutil.NewBigEndianStream(p)

	if _, err := binaryutil.CopyToByteStream(&bs, v.TypeId); err != nil {
		return bs.BytesWritten(), err
	}

	if v.TypeId >= TypeId_Customize {
		if err := bs.WriteUvarint(uint64(v.Value.Size())); err != nil {
			return bs.BytesWritten(), err
		}
	}

	if _, err := binaryutil.CopyToByteStream(&bs, v.Value); err != nil {
		return bs.BytesWritten(), err
	}

	return bs.BytesWritten(), io.EOF
}

// Write implements io.Writer
func (v *Variant) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if _, err := bs.WriteTo(&v.TypeId); err != nil {
		return bs.BytesRead(), err
	}

	if v.TypeId >= TypeId_Customize {
		valueSize, err := bs.ReadUvarint()
		if err != nil {
			return bs.BytesRead(), err
		}
		if uint64(bs.BytesUnread()) < valueSize {
			return bs.BytesRead(), io.ErrUnexpectedEOF
		}
	}

	reflected, err := v.TypeId.NewReflected()
	if err != nil {
		return bs.BytesRead(), err
	}

	value := reflected.Interface().(Value)
	if _, err := bs.WriteTo(value); err != nil {
		return bs.BytesRead(), err
	}

	v.Value = value
	v.Reflected = reflected

	return bs.BytesRead(), nil
}

// Size 大小
func (v Variant) Size() int {
	if !v.IsValid() {
		return 0
	}

	n := v.TypeId.Size()

	if v.Value != nil {
		s := v.Value.Size()
		if v.TypeId >= TypeId_Customize {
			n += binaryutil.SizeofUvarint(uint64(s))
		}
		n += s
	}

	return n
}

// IsValid 是否有效
func (v Variant) IsValid() bool {
	if v.Value != nil {
		return v.TypeId == v.Value.TypeId()
	}
	return false
}
