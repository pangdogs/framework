package gap

import (
	"kit.golaxy.org/plugins/gap/variant"
	"kit.golaxy.org/plugins/util/binaryutil"
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

	rn := bs.BytesWritten()

	n, err := m.Rets.Read(p[rn:])
	rn += n
	if err != nil {
		return rn, nil
	}

	n, err = m.Error.Read(p[rn:])
	rn += n
	if err != nil {
		return rn, nil
	}

	return rn, nil
}

// Write implements io.Writer
func (m *MsgRPCReply) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	corrId, err := bs.ReadVarint()
	if err != nil {
		return bs.BytesRead(), err
	}

	wn := bs.BytesRead()

	var rets variant.Array
	n, err := rets.Write(p[wn:])
	wn += n
	if err != nil {
		return wn, err
	}

	var retErr variant.Error
	n, err = retErr.Write(p[wn:])
	wn += n
	if err != nil {
		return wn, err
	}

	m.CorrId = corrId
	m.Rets = rets
	m.Error = retErr

	return wn, nil
}

// Size 大小
func (m MsgRPCReply) Size() int {
	return binaryutil.SizeofVarint(m.CorrId) + m.Rets.Size() + m.Error.Size()
}

// MsgId 消息Id
func (MsgRPCReply) MsgId() MsgId {
	return MsgId_RPC_Reply
}
