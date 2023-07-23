package transport

import "kit.golaxy.org/plugins/transport/binaryutil"

// MsgPayload 数据传输（注意：为了提高解码性能，减少内存碎片，解码string与bytes字段时均使用引用类型，引用字节池中的bytes，GC时会被归还字节池，不要直接持有此类型字段）
type MsgPayload struct {
	Data []byte // 数据
}

func (m *MsgPayload) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteBytes(m.Data); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (m *MsgPayload) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	data, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	m.Data = data
	return bs.BytesRead(), nil
}

func (m *MsgPayload) Size() int {
	return binaryutil.SizeofBytes(m.Data)
}

func (MsgPayload) MsgId() MsgId {
	return MsgId_Payload
}
