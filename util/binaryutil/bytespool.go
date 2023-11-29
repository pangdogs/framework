package binaryutil

import (
	"github.com/fufuok/bytespool"
	"math"
)

// BytesPool 字节对象池，用于减少GC提高编解码性能
var BytesPool = bytespool.NewCapacityPools(32, math.MaxInt32)

// MakeRecycleBytes 创建可回收字节对象
func MakeRecycleBytes(bytes []byte) RecycleBytes {
	return RecycleBytes{
		data:       bytes,
		low:        0,
		high:       len(bytes),
		recyclable: true,
	}
}

// MakeNonRecycleBytes 创建不可回收字节对象
func MakeNonRecycleBytes(bytes []byte) RecycleBytes {
	return RecycleBytes{
		data:       bytes,
		low:        0,
		high:       len(bytes),
		recyclable: false,
	}
}

// SliceRecycleBytes 切片操作字节对象
func SliceRecycleBytes(bytes RecycleBytes, low, high int) RecycleBytes {
	if low < 0 || high < 0 {
		panic("negative index")
	}
	if high == 0 {
		high = len(bytes.data) - bytes.low
	}
	return RecycleBytes{
		data:       bytes.data,
		low:        bytes.low + low,
		high:       bytes.low + high,
		recyclable: bytes.recyclable,
	}
}

// RecycleBytes 可回收字节对象
type RecycleBytes struct {
	data       []byte
	low, high  int
	recyclable bool
}

// Data 数据
func (b RecycleBytes) Data() []byte {
	return b.data[b.low:b.high]
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
