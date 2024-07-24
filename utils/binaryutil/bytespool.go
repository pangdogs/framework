package binaryutil

import (
	"github.com/fufuok/bytespool"
	"math"
	"reflect"
	"unsafe"
)

// BytesPool 字节对象池，用于减少GC提高编解码性能
var BytesPool = bytespool.NewCapacityPools(32, math.MaxInt32)

// NilRecycleBytes 空字节对象
var NilRecycleBytes = MakeNonRecycleBytes(nil)

// MakeRecycleBytes 创建可回收字节对象
func MakeRecycleBytes(size int) RecycleBytes {
	return RecycleBytes{
		data:       BytesPool.Get(size),
		low:        0,
		high:       size,
		recyclable: true,
	}
}

// CloneRecycleBytes 克隆并创建可回收字节对象
func CloneRecycleBytes(bs []byte) RecycleBytes {
	return RecycleBytes{
		data:       BytesPool.Clone(bs),
		low:        0,
		high:       len(bs),
		recyclable: true,
	}
}

// MakeNonRecycleBytes 创建不可回收字节对象
func MakeNonRecycleBytes(bs []byte) RecycleBytes {
	return RecycleBytes{
		data:       bs,
		low:        0,
		high:       len(bs),
		recyclable: false,
	}
}

// RecycleBytes 可回收字节对象
type RecycleBytes struct {
	data       []byte
	low, high  int
	recyclable bool
}

// Equal 是否是相同
func (b RecycleBytes) Equal(other RecycleBytes) bool {
	refA := (*reflect.SliceHeader)(unsafe.Pointer(&b.data)).Data
	refB := (*reflect.SliceHeader)(unsafe.Pointer(&other.data)).Data
	return refA == refB
}

// Data 数据
func (b RecycleBytes) Data() []byte {
	return b.data[b.low:b.high]
}

// Slice 切片
func (b RecycleBytes) Slice(low, high int) RecycleBytes {
	if low < 0 || high < 0 {
		panic("negative index")
	}
	if high == 0 {
		high = len(b.data) - b.low
	}
	return RecycleBytes{
		data:       b.data,
		low:        b.low + low,
		high:       b.low + high,
		recyclable: b.recyclable,
	}
}

// Recyclable 可否回收
func (b RecycleBytes) Recyclable() bool {
	return b.recyclable
}

// Release 释放字节对象
func (b RecycleBytes) Release() {
	if b.recyclable {
		BytesPool.Put(b.data)
	}
}
