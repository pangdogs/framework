package gap

import "io"

// MsgId 消息Id
type MsgId = uint32

const (
	MsgId_None        MsgId  = iota // 未设置
	MsgId_RPC_Request               // RPC请求
	MsgId_RPC_Reply                 // RPC响应
	MsgId_Notify                    // 通知
	MsgId_Customize   = 1024        // 自定义消息起点
)

// Msg 消息接口
type Msg interface {
	io.ReadWriter
	// Size 大小
	Size() int
	// MsgId 消息Id
	MsgId() MsgId
}
