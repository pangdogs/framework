package gtp

// Heartbeat消息标志位
const (
	Flag_Ping Flag = 1 << (iota + Flag_Customize) // 心跳ping
	Flag_Pong                                     // 心跳pong
)

// MsgHeartbeat 心跳，消息体为空，可以不解析
type MsgHeartbeat struct{}

// Read implements io.Reader
func (MsgHeartbeat) Read(p []byte) (int, error) {
	return 0, nil
}

// Write implements io.Writer
func (MsgHeartbeat) Write(p []byte) (int, error) {
	return 0, nil
}

// Size 消息大小
func (MsgHeartbeat) Size() int {
	return 0
}

// MsgId 消息Id
func (MsgHeartbeat) MsgId() MsgId {
	return MsgId_Heartbeat
}

// Clone 克隆消息对象
func (m MsgHeartbeat) Clone() Msg {
	return &m
}
