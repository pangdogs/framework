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
	// ErrVariant 是 GAP 动态值处理错误的根错误。
	ErrVariant = errors.New("gap-variant")
)

// NewVariant 用值自身的类型 ID 创建动态值包装。
func NewVariant(v ReadableValue) (Variant, error) {
	if v == nil {
		return Variant{}, fmt.Errorf("%w: %w: v is nil", ErrVariant, core.ErrArgs)
	}
	return Variant{
		TypeID: v.TypeID(),
		Value:  v,
	}, nil
}

// Variant 将类型 ID 与可编码值关联，并可保留解码时创建的反射值。
type Variant struct {
	TypeID    TypeID        // 动态值类型 ID。
	Value     ReadableValue // 实际值。
	Reflected reflect.Value // 解码自定义类型时创建的反射值；自行构造时可为空。
}

// Read 将带类型信息的动态值编码到 p。
func (v Variant) Read(p []byte) (int, error) {
	if !v.IsValid() {
		return 0, fmt.Errorf("%w: invalid variant", ErrVariant)
	}

	bs := binaryutil.NewBigEndianStream(p)

	if _, err := binaryutil.CopyToByteStream(&bs, v.TypeID); err != nil {
		return bs.BytesWritten(), err
	}

	if v.TypeID >= TypeID_Customize {
		if err := bs.WriteUvarint(uint64(v.Value.Size())); err != nil {
			return bs.BytesWritten(), err
		}
	}

	if _, err := binaryutil.CopyToByteStream(&bs, v.Value); err != nil {
		return bs.BytesWritten(), err
	}

	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码动态值，并通过 VariantCreator 构造其具体类型。
func (v *Variant) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if _, err := bs.WriteTo(&v.TypeID); err != nil {
		return bs.BytesRead(), err
	}

	if v.TypeID >= TypeID_Customize {
		valueSize, err := bs.ReadUvarint()
		if err != nil {
			return bs.BytesRead(), err
		}
		if uint64(bs.BytesUnread()) < valueSize {
			return bs.BytesRead(), io.ErrUnexpectedEOF
		}
	}

	reflected, err := v.TypeID.NewReflected()
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

// Size 返回动态值连同类型信息编码后的字节数；无效值返回零。
func (v Variant) Size() int {
	if !v.IsValid() {
		return 0
	}

	n := v.TypeID.Size()

	if v.Value != nil {
		s := v.Value.Size()
		if v.TypeID >= TypeID_Customize {
			n += binaryutil.SizeofUvarint(uint64(s))
		}
		n += s
	}

	return n
}

// IsValid 报告类型 ID 是否与实际值声明的类型一致。
func (v Variant) IsValid() bool {
	if v.Value != nil {
		return v.TypeID == v.Value.TypeID()
	}
	return false
}
