package transport

// MsgId 消息Id
type MsgId = uint8

const (
	MsgId_Hello             MsgId = iota // Hello Handshake C<->S 不加密
	MsgId_SecretKeyExchange              // 秘钥交换 Handshake S<->C 不加密
	MsgId_ChangeCipherSpec               // 变更密码规范 Handshake S<->C 不加密
	MsgId_Auth                           // 鉴权 Handshake C->S 加密
	MsgId_Finished                       // 握手结束 Handshake S<->C 加密
	MsgId_Rst                            // 重置链路 Ctrl S->C 加密
	MsgId_Heartbeat                      // 心跳 Ctrl C<->S or S<->C 加密
	MsgId_SyncTime                       // 时钟同步 Ctrl S->C 加密
	MsgId_Payload                        // 数据传输 Trans C<->S or S<->C 加密
)
