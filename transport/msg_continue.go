package transport

import "kit.golaxy.org/plugins/transport/binaryutil"

// MsgContinue 重连
type MsgContinue struct {
	SendSeq uint32 // 客户端请求消息序号
	RecvSeq uint32 // 客户端响应消息序号
}

func (m *MsgContinue) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteUint32(m.SendSeq); err != nil {
		return 0, err
	}
	if err := bs.WriteUint32(m.RecvSeq); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (m *MsgContinue) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	sendSeq, err := bs.ReadUint32()
	if err != nil {
		return 0, err
	}
	recvSeq, err := bs.ReadUint32()
	if err != nil {
		return 0, err
	}
	m.SendSeq = sendSeq
	m.RecvSeq = recvSeq
	return bs.BytesRead(), nil
}

func (m *MsgContinue) Size() int {
	return binaryutil.SizeofUint32() + binaryutil.SizeofUint32() + binaryutil.SizeofUint32()
}

func (MsgContinue) MsgId() MsgId {
	return MsgId_Continue
}
