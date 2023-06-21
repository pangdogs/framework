package transport

// Heartbeat消息标志位
const (
	Flag_Ping Flag = 1 << (iota + Flag_Customize) // 心跳ping
	Flag_Pong                                     // 心跳pong
)

// MsgHeartbeat 心跳，消息体为空，可以不解析
type MsgHeartbeat struct{}

func (MsgHeartbeat) Read(p []byte) (int, error) {
	return 0, nil
}

func (MsgHeartbeat) Write(p []byte) (int, error) {
	return 0, nil
}

func (MsgHeartbeat) Size() int {
	return 0
}

func (MsgHeartbeat) MsgId() MsgId {
	return MsgId_Heartbeat
}
