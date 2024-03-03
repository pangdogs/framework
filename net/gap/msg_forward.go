package gap

import (
	"git.golaxy.org/framework/util/binaryutil"
)

// MsgForward 转发
type MsgForward struct {
	Dst string // 转发目标
	Raw []byte // 原始消息（引用）
}

// Read implements io.Reader
func (m MsgForward) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteString(m.Dst); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteBytes(m.Raw); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgForward) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.Dst, err = bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.Raw, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgForward) Size() int {
	return binaryutil.SizeofString(m.Dst) + binaryutil.SizeofBytes(m.Raw)
}

// MsgId 消息Id
func (MsgForward) MsgId() MsgId {
	return MsgId_Forward
}
