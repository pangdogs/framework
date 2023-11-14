package gap

import (
	"kit.golaxy.org/plugins/util/binaryutil"
)

// MsgHead 消息头
type MsgHead struct {
	MsgId   MsgId  // 消息Id
	SeqId   int64  // 序号
	Address string // 服务节点地址
}

// Read implements io.Reader
func (m *MsgHead) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteVarint(int64(m.MsgId)); err != nil {
		return 0, err
	}
	if err := bs.WriteVarint(m.SeqId); err != nil {
		return 0, err
	}
	if err := bs.WriteString(m.Address); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgHead) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	msgId, err := bs.ReadVarint()
	if err != nil {
		return 0, err
	}
	seqId, err := bs.ReadVarint()
	if err != nil {
		return 0, err
	}
	address, err := bs.ReadString()
	if err != nil {
		return 0, err
	}
	m.MsgId = MsgId(msgId)
	m.SeqId = seqId
	m.Address = address
	return bs.BytesRead(), nil
}

// Size 大小
func (m *MsgHead) Size() int {
	return binaryutil.SizeofVarint(int64(m.MsgId)) + binaryutil.SizeofVarint(m.SeqId) + binaryutil.SizeofString(m.Address)
}
