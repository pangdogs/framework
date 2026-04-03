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

package binaryutil

import (
	"bytes"
	"math"
	"reflect"
	"unsafe"

	"git.golaxy.org/core/utils/exception"
	"github.com/fufuok/bytespool"
)

// BytesPool 字节对象池，用于减少GC提高编解码性能
var BytesPool = bytespool.NewCapacityPools(32, math.MaxInt32)

// EmptyBytes 空字节对象
var EmptyBytes = NewBytes(false, 0)

// NewBytes 创建字节对象
func NewBytes(recyclable bool, size int) Bytes {
	if size < 0 {
		size = 0
	}
	bs := Bytes{
		low:        0,
		high:       size,
		recyclable: recyclable,
	}
	if recyclable {
		bs.data = BytesPool.Get(size)
	} else {
		bs.data = make([]byte, size)
	}
	return bs
}

// CloneBytes 克隆并创建字节对象
func CloneBytes(recyclable bool, buff []byte) Bytes {
	bs := Bytes{
		low:        0,
		high:       len(buff),
		recyclable: recyclable,
	}
	if recyclable {
		bs.data = BytesPool.Clone(buff)
	} else {
		bs.data = bytes.Clone(buff)
	}
	return bs
}

// RefBytes 引用并创建字节对象（不能回收）
func RefBytes(buff []byte) Bytes {
	return Bytes{
		data:       buff,
		low:        0,
		high:       len(buff),
		recyclable: false,
	}
}

// Bytes 字节对象
type Bytes struct {
	data       []byte
	low, high  int
	recyclable bool
}

// SameRef 是否引用相同数据
func (bs Bytes) SameRef(other Bytes) bool {
	refA := (*reflect.SliceHeader)(unsafe.Pointer(&bs.data)).Data
	refB := (*reflect.SliceHeader)(unsafe.Pointer(&other.data)).Data
	return refA == refB
}

// Payload 载荷数据
func (bs Bytes) Payload() []byte {
	return bs.data[bs.low:bs.high]
}

// Slice 切片
func (bs Bytes) Slice(low, high int) Bytes {
	if low < 0 || high < 0 {
		exception.Panic("negative index")
	}
	if low > high {
		exception.Panic("low > high")
	}
	if bs.low+high > bs.high {
		exception.Panic("slice out of range")
	}
	return Bytes{
		data:       bs.data,
		low:        bs.low + low,
		high:       bs.low + high,
		recyclable: bs.recyclable,
	}
}

// Recyclable 可否回收
func (bs Bytes) Recyclable() bool {
	return bs.recyclable
}

// Release 释放字节对象，释放后不可再使用，不能重复释放
func (bs Bytes) Release() {
	if bs.recyclable {
		BytesPool.Put(bs.data)
	}
}
