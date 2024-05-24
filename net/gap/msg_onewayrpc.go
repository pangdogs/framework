package gap

import (
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/util/binaryutil"
)

// MsgOneWayRPC 单程RPC请求
type MsgOneWayRPC struct {
	CallChain variant.CallChain // 调用链
	Path      string            // 调用路径
	Args      variant.Array     // 参数列表
}

// Read implements io.Reader
func (m MsgOneWayRPC) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if _, err := binaryutil.ReadFrom(&bs, m.CallChain); err != nil {
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
func (m *MsgOneWayRPC) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	if _, err = bs.WriteTo(&m.CallChain); err != nil {
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
func (m MsgOneWayRPC) Size() int {
	return m.CallChain.Size() + binaryutil.SizeofString(m.Path) + m.Args.Size()
}

// MsgId 消息Id
func (MsgOneWayRPC) MsgId() MsgId {
	return MsgId_OneWayRPC
}
