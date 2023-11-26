package variant

import (
	"io"
)

// Value 值
type Value interface {
	io.ReadWriter
	// Size 大小
	Size() int
	// Type 类型
	Type() TypeId
}
