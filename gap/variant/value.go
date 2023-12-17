package variant

import (
	"io"
)

// Value 值
type Value interface {
	ValueReader
	ValueWriter
}

// ValueReader 读取值
type ValueReader interface {
	io.Reader
	// Size 大小
	Size() int
	// Type 类型
	Type() TypeId
}

// ValueWriter 写入值
type ValueWriter interface {
	io.Writer
	// Size 大小
	Size() int
	// Type 类型
	Type() TypeId
}
