package binaryutil

import (
	"github.com/fufuok/bytespool"
	"math"
)

// BytesPool 字节对象池，用于减少GC提高编解码性能
var BytesPool = bytespool.NewCapacityPools(32, math.MaxInt32)
