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

// NewSerializedArray 创建已序列化array
func NewSerializedArray[T any](arr []T) (ret SerializedArray, err error) {
	ret.Array = make(Array, 0, len(arr))
	ret.items = make([]SerializedVariant, 0, len(arr))
	defer func() {
		if err != nil {
			ret.Release()
			ret = SerializedArray{}
		}
	}()

	for i := range arr {
		v, err := CastSerializedVariant(&arr[i])
		if err != nil {
			return SerializedArray{}, err
		}
		ret.Array = append(ret.Array, v.Ref())
		ret.items = append(ret.items, v)
	}

	return ret, nil
}

// SerializedArray 已序列化数组
type SerializedArray struct {
	Array
	items []SerializedVariant
}

// Release 释放缓存，缓存释放后请勿再使用或引用变体值
func (a SerializedArray) Release() {
	for i := range a.items {
		a.items[i].Release()
	}
}

// Ref 引用变体值
func (a SerializedArray) Ref() Array {
	return a.Array
}
