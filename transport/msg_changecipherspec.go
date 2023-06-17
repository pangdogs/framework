package transport

import "kit.golaxy.org/plugins/transport/binaryutil"

// ChangeCipherSpec消息标志位
const (
	Flag_VerifyEncryption Flag = 1 << (iota + Flag_Customize) // 交换秘钥后，在服务端变更密码规范消息中携带，表示需要验证加密是否成功
)

// MsgChangeCipherSpec 变更密码规范
type MsgChangeCipherSpec struct {
	EncryptedHello []byte // 加密Hello消息，用于双方验证加密是否成功
}

func (m *MsgChangeCipherSpec) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteBytes(m.EncryptedHello); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (m *MsgChangeCipherSpec) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	encryptedHello, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	m.EncryptedHello = encryptedHello
	return bs.BytesRead(), nil
}

func (m *MsgChangeCipherSpec) Size() int {
	return binaryutil.SizeofBytes(m.EncryptedHello)
}

func (MsgChangeCipherSpec) MsgId() MsgId {
	return MsgId_ChangeCipherSpec
}
