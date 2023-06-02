package transport

import "kit.golaxy.org/plugins/transport/binaryutil"

// MsgChangeCipherSpec 变更密码规范
type MsgChangeCipherSpec struct {
	EncryptedHelloMsg []byte // 加密Hello消息，双方验证加密是否成功
}

func (m *MsgChangeCipherSpec) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	encryptedHelloMsg, err := bs.ReadBytes()
	if err != nil {
		return 0, err
	}
	m.EncryptedHelloMsg = encryptedHelloMsg
	return bs.BytesRead(), nil
}

func (m *MsgChangeCipherSpec) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteBytes(m.EncryptedHelloMsg); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (m *MsgChangeCipherSpec) Size() int {
	return binaryutil.SizeofBytes(m.EncryptedHelloMsg)
}
