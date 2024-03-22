package gap

import (
	"git.golaxy.org/framework/util/binaryutil"
)

// MsgForward 转发
type MsgForward struct {
	Gate      string // 网关地址
	Dst       string // 目标地址
	CorrId    int64  // 关联Id，用于支持Future等异步模型
	TransId   MsgId  // 传输消息Id
	TransData []byte // 传输消息内容（引用）
}

// Read implements io.Reader
func (m MsgForward) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteString(m.Gate); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(m.Dst); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteVarint(m.CorrId); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint32(m.TransId); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteBytes(m.TransData); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgForward) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.Gate, err = bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.Dst, err = bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.CorrId, err = bs.ReadVarint()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.TransId, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.TransData, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgForward) Size() int {
	return binaryutil.SizeofString(m.Gate) + binaryutil.SizeofString(m.Dst) + binaryutil.SizeofVarint(m.CorrId) + binaryutil.SizeofUint32() + binaryutil.SizeofBytes(m.TransData)
}

// MsgId 消息Id
func (MsgForward) MsgId() MsgId {
	return MsgId_Forward
}
