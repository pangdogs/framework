package gap

import (
	"io"
)

// SerializedMsg 已序列化消息
type SerializedMsg struct {
	Id   MsgId  // 消息Id
	Data []byte // 消息内容（引用）
}

// Read implements io.Reader
func (m *SerializedMsg) Read(p []byte) (int, error) {
	if len(p) < len(m.Data) {
		return 0, io.ErrShortWrite
	}
	return copy(p, m.Data), nil
}

// Size 大小
func (m *SerializedMsg) Size() int {
	return len(m.Data)
}

// MsgId 消息Id
func (m *SerializedMsg) MsgId() MsgId {
	return m.Id
}
