package gap

import (
	"kit.golaxy.org/plugins/gap/variant"
	"kit.golaxy.org/plugins/util/binaryutil"
)

type MsgOneWayRPC struct {
	Path string        // 调用路径
	Args variant.Array // 参数列表
}

// Read implements io.Reader
func (m MsgOneWayRPC) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
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
	return binaryutil.SizeofString(m.Path) + m.Args.Size()
}

// MsgId 消息Id
func (MsgOneWayRPC) MsgId() MsgId {
	return MsgId_OneWayRPC
}
