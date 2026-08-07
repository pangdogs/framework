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

// BytesPool 按容量复用字节切片，以降低协议编解码产生的 GC 压力。
var BytesPool = bytespool.NewCapacityPools(32, math.MaxInt32)

// EmptyBytes 是不可回收的空字节缓冲区。
var EmptyBytes = NewBytes(false, 0)

// NewBytes 创建长度为 size 的零值字节缓冲区；负数 size 按零处理。
// recyclable 为 true 时缓冲区取自 BytesPool，调用方使用完后必须调用 Release。
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

// CloneBytes 复制 buff 并创建独立缓冲区。
// recyclable 为 true 时缓冲区取自 BytesPool，调用方使用完后必须调用 Release。
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

// RefBytes 创建直接引用 buff 的不可回收缓冲区，不复制数据。
func RefBytes(buff []byte) Bytes {
	return Bytes{
		data:       buff,
		low:        0,
		high:       len(buff),
		recyclable: false,
	}
}

// Bytes 表示一个可切片、可选池化的字节缓冲区视图。
//
// Bytes 及其 Slice 结果可能共享底层存储。对于可回收缓冲区，所有共享视图只能由其所有者
// 调用一次 Release；释放后所有视图及 Payload 返回的切片均不得再使用。
type Bytes struct {
	data       []byte
	low, high  int
	recyclable bool
}

// SameRef 报告两个缓冲区是否具有相同的底层切片起始地址。
func (bs Bytes) SameRef(other Bytes) bool {
	refA := (*reflect.SliceHeader)(unsafe.Pointer(&bs.data)).Data
	refB := (*reflect.SliceHeader)(unsafe.Pointer(&other.data)).Data
	return refA == refB
}

// Payload 返回当前视图的可修改切片；返回值与 Bytes 共享底层存储。
func (bs Bytes) Payload() []byte {
	return bs.data[bs.low:bs.high]
}

// Slice 返回当前视图区间 [low, high) 的共享视图。
// 索引越界或区间无效时 panic；返回值不会获得独立的 Release 责任。
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

// Recyclable 报告底层切片是否应归还 BytesPool。
func (bs Bytes) Recyclable() bool {
	return bs.recyclable
}

// Release 将可回收缓冲区归还 BytesPool；不可回收缓冲区调用此方法无效果。
// 可回收缓冲区释放后及其所有共享视图均不得再使用，也不得重复释放。
func (bs Bytes) Release() {
	if bs.recyclable {
		BytesPool.Put(bs.data)
	}
}
