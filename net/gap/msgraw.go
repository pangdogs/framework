package gap

import "git.golaxy.org/framework/util/binaryutil"

// MsgRaw 原始消息
type MsgRaw struct {
	Id   MsgId  // 消息Id
	Data []byte // 消息内容（引用）
}

// Read implements io.Reader
func (m MsgRaw) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteBytes(m.Data); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgRaw) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.Data, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgRaw) Size() int {
	return binaryutil.SizeofUint32() + binaryutil.SizeofBytes(m.Data)
}

// MsgId 消息Id
func (m MsgRaw) MsgId() MsgId {
	return m.Id
}
