package gap

import (
	"kit.golaxy.org/plugins/gap/variant"
	"kit.golaxy.org/plugins/util/binaryutil"
)

// MsgRPCRequest RPC请求
type MsgRPCRequest struct {
	CorrId int64         // 关联Id，用于支持Future等异步模型
	Path   string        // 调用路径
	Args   variant.Array // 参数列表
}

// Read implements io.Reader
func (m MsgRPCRequest) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteVarint(m.CorrId); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(m.Path); err != nil {
		return bs.BytesWritten(), err
	}
	if _, err := binaryutil.ReadFrom(&bs, m.Args); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgRPCRequest) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.CorrId, err = bs.ReadVarint()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.Path, err = bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}

	if _, err = bs.WriteTo(&m.Args); err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgRPCRequest) Size() int {
	return binaryutil.SizeofVarint(m.CorrId) + binaryutil.SizeofString(m.Path) + m.Args.Size()
}

// MsgId 消息Id
func (MsgRPCRequest) MsgId() MsgId {
	return MsgId_RPC_Request
}
