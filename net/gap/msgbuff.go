package gap

import (
	"io"
)

// MsgBuff msg buff
type MsgBuff struct {
	Id   MsgId  // 消息Id
	Data []byte // 消息内容（引用）
}

// Read implements io.Reader
func (m MsgBuff) Read(p []byte) (int, error) {
	if len(p) < len(m.Data) {
		return 0, io.ErrShortWrite
	}
	return copy(p, m.Data), nil
}

// Write implements io.Writer
func (m *MsgBuff) Write(p []byte) (int, error) {
	m.Data = p
	return len(p), nil
}

// Size 大小
func (m MsgBuff) Size() int {
	return len(m.Data)
}

// MsgId 消息Id
func (m MsgBuff) MsgId() MsgId {
	return m.Id
}
