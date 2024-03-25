package gap

import (
	"io"
)

// MsgRaw 原始消息
type MsgRaw struct {
	Id   MsgId  // 消息Id
	Data []byte // 消息内容（引用）
}

// Read implements io.Reader
func (m MsgRaw) Read(p []byte) (int, error) {
	if len(p) < len(m.Data) {
		return 0, io.ErrShortWrite
	}
	return copy(p, m.Data), nil
}

// Write implements io.Writer
func (m *MsgRaw) Write(p []byte) (int, error) {
	m.Data = p
	return len(p), nil
}

// Size 大小
func (m MsgRaw) Size() int {
	return len(m.Data)
}

// MsgId 消息Id
func (m MsgRaw) MsgId() MsgId {
	return m.Id
}
