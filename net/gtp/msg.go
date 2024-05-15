package gtp

import "io"

// Msg 消息接口
type Msg interface {
	MsgReader
	MsgWriter
	// Clone 克隆消息对象
	Clone() Msg
}

// MsgReader 读取消息
type MsgReader interface {
	io.Reader
	// Size 大小
	Size() int
	// MsgId 消息Id
	MsgId() MsgId
}

// MsgWriter 写入消息
type MsgWriter interface {
	io.Writer
	// Size 大小
	Size() int
	// MsgId 消息Id
	MsgId() MsgId
}
