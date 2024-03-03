package gap

import "io"

// MsgId 消息Id
type MsgId = uint32

const (
	MsgId_None        MsgId = iota // 未设置
	MsgId_RPC_Request              // RPC请求
	MsgId_RPC_Reply                // RPC答复
	MsgId_OneWayRPC                // 单程RPC请求
	MsgId_Forward                  // 转发
	MsgId_Customize   = 128        // 自定义消息起点
)

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
