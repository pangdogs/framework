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
)

// NewSerializedVariant 创建已序列化变体
func NewSerializedVariant(v ReadableValue) (SerializedVariant, error) {
	if v == nil {
		return SerializedVariant{}, fmt.Errorf("%w: %w: v is nil", ErrVariant, core.ErrArgs)
	}

	sv, err := NewSerializedValue(v)
	if err != nil {
		return SerializedVariant{}, err
	}

	return wrappedSerializedVariant(sv, sv), nil
}

func wrappedSerializedVariant(v ReadableValue, r releasable) SerializedVariant {
	return SerializedVariant{
		Variant: Variant{
			TypeId: v.TypeId(),
			Value:  v,
		},
		releasable: r,
	}
}

type releasable interface {
	Release()
}

// SerializedVariant 已序列化变体
type SerializedVariant struct {
	Variant
	releasable releasable
}

// Release 释放缓存，缓存释放后请勿再使用或引用变体
func (v SerializedVariant) Release() {
	if v.releasable != nil {
		v.releasable.Release()
	}
}

// Ref 引用变体
func (v SerializedVariant) Ref() Variant {
	return v.Variant
}
