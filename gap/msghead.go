package gap

import (
	"kit.golaxy.org/plugins/util/binaryutil"
)

// MsgHead 消息头
type MsgHead struct {
	Len   uint32 // 消息长度
	MsgId MsgId  // 消息Id
	Src   string // 来源地址
	Seq   int64  // 序号
}

// Read implements io.Reader
func (m MsgHead) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint32(m.Len); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint32(m.MsgId); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(m.Src); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteInt64(m.Seq); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgHead) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	l, err := bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}
	msgId, err := bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}
	src, err := bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}
	seq, err := bs.ReadInt64()
	if err != nil {
		return bs.BytesRead(), err
	}
	m.Len = l
	m.MsgId = msgId
	m.Src = src
	m.Seq = seq
	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgHead) Size() int {
	return binaryutil.SizeofUint32() +
		binaryutil.SizeofUint32() +
		binaryutil.SizeofString(m.Src) +
		binaryutil.SizeofInt64()
}
