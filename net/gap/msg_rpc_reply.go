package gap

import (
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/util/binaryutil"
)

// MsgRPCReply RPC答复
type MsgRPCReply struct {
	CorrId int64         // 关联Id，用于支持Future等异步模型
	Rets   variant.Array // 调用结果
	Error  variant.Error // 调用错误
}

// Read implements io.Reader
func (m MsgRPCReply) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteVarint(m.CorrId); err != nil {
		return bs.BytesWritten(), err
	}
	if _, err := binaryutil.ReadFrom(&bs, m.Rets); err != nil {
		return bs.BytesWritten(), err
	}
	if _, err := binaryutil.ReadFrom(&bs, m.Error); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgRPCReply) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.CorrId, err = bs.ReadVarint()
	if err != nil {
		return bs.BytesRead(), err
	}

	if _, err = bs.WriteTo(&m.Rets); err != nil {
		return bs.BytesRead(), err
	}

	if _, err = bs.WriteTo(&m.Error); err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgRPCReply) Size() int {
	return binaryutil.SizeofVarint(m.CorrId) + m.Rets.Size() + m.Error.Size()
}

// MsgId 消息Id
func (MsgRPCReply) MsgId() MsgId {
	return MsgId_RPC_Reply
}
