package transport

import "kit.golaxy.org/plugins/transport/binaryutil"

// Finished消息标志位
const (
	Flag_EncryptOK  Flag = 1 << (iota + Flag_Customize) // 加密成功，在服务端发起的Finished消息携带
	Flag_AuthOK                                         // 鉴权成功，在服务端发起的Finished消息携带
	Flag_ContinueOK                                     // 断线重连成功，在服务端发起的Finished消息携带
)

// MsgFinished 握手结束，表示认可对端，可以开始传输数据
type MsgFinished struct {
	Seq uint32 // 消息序号
	Ack uint32 // 应答序号
}

func (m *MsgFinished) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteUint32(m.Seq); err != nil {
		return 0, err
	}
	if err := bs.WriteUint32(m.Ack); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (m *MsgFinished) Write(p []byte) (int, error) {
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

func (m *MsgFinished) Size() int {
	return binaryutil.SizeofUint32() + binaryutil.SizeofUint32()
}

func (MsgFinished) MsgId() MsgId {
	return MsgId_Finished
}
