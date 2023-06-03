package transport

// MsgHeartbeat 心跳，消息体为空，可以不解析
type MsgHeartbeat struct{}

func (m *MsgHeartbeat) Read(p []byte) (int, error) {
	return 0, nil
}

func (m *MsgHeartbeat) Write(p []byte) (int, error) {
	return 0, nil
}

func (m *MsgHeartbeat) Size() int {
	return 0
}
