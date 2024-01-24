package gap

import (
	"git.golaxy.org/framework/plugins/util/binaryutil"
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
	var err error

	m.Len, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.MsgId, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.Src, err = bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.Seq, err = bs.ReadInt64()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgHead) Size() int {
	return binaryutil.SizeofUint32() + binaryutil.SizeofUint32() + binaryutil.SizeofString(m.Src) +
		binaryutil.SizeofInt64()
}
