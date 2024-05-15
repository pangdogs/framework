package gap

const (
	MsgId_None        MsgId = iota // 未设置
	MsgId_RPC_Request              // RPC请求
	MsgId_RPC_Reply                // RPC答复
	MsgId_OneWayRPC                // 单程RPC请求
	MsgId_Forward                  // 转发
	MsgId_Customize   = 32         // 自定义消息起点
)
