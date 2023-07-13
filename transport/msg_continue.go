package transport

import "kit.golaxy.org/plugins/transport/binaryutil"

// MsgContinue 重连
type MsgContinue struct {
	Seq uint32 // 消息序号
	Ack uint32 // 应答序号
}

func (m *MsgContinue) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteUint32(m.Seq); err != nil {
		return 0, err
	}
	if err := bs.WriteUint32(m.Ack); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (m *MsgContinue) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	seq, err := bs.ReadUint32()
	if err != nil {
		return 0, err
	}
	ack, err := bs.ReadUint32()
	if err != nil {
		return 0, err
	}
	m.Seq = seq
	m.Ack = ack
	return bs.BytesRead(), nil
}

func (m *MsgContinue) Size() int {
	return binaryutil.SizeofUint32() + binaryutil.SizeofUint32()
}

func (MsgContinue) MsgId() MsgId {
	return MsgId_Continue
}
