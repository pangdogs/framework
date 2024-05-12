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
	// TypeId 类型
	TypeId() TypeId
	// Indirect 原始值
	Indirect() any
}

// ValueWriter 写入值
type ValueWriter interface {
	io.Writer
	// Size 大小
	Size() int
	// TypeId 类型
	TypeId() TypeId
}
