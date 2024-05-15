package gap

import "io"

// Msg 消息接口
type Msg interface {
	MsgReader
	MsgWriter
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
