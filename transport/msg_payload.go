package transport

import "kit.golaxy.org/plugins/transport/binaryutil"

// MsgPayload 数据传输
type MsgPayload struct {
	Seq  uint32 // 消息序号
	Data []byte // 数据
}

func (m *MsgPayload) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	seq, err := bs.ReadUint32()
	if err != nil {
		return 0, err
	}
	data, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	m.Seq = seq
	m.Data = data
	return bs.BytesRead(), nil
}

func (m *MsgPayload) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteUint32(m.Seq); err != nil {
		return 0, err
	}
	if err := bs.WriteBytes(m.Data); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (m *MsgPayload) Size() int {
	return binaryutil.SizeofUint32() + binaryutil.SizeofBytes(m.Data)
}

func (MsgPayload) MsgId() MsgId {
	return MsgId_Payload
}
