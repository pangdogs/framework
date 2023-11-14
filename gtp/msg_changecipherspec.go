package gtp

import (
	"bytes"
	"kit.golaxy.org/plugins/util/binaryutil"
)

// MsgChangeCipherSpec消息标志位
const (
	Flag_VerifyEncryption Flag = 1 << (iota + Flag_Customize) // 交换秘钥后，在双方变更密码规范消息中携带，表示需要验证加密是否成功
)

// MsgChangeCipherSpec 变更密码规范（注意：为了提高解码性能，减少内存碎片，解码string与bytes字段时均使用引用类型，引用字节池中的bytes，GC时会被归还字节池，不要直接持有此类型字段）
type MsgChangeCipherSpec struct {
	EncryptedHello []byte // 加密Hello消息，用于双方验证加密是否成功
}

// Read implements io.Reader
func (m *MsgChangeCipherSpec) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteBytes(m.EncryptedHello); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgChangeCipherSpec) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	encryptedHello, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	m.EncryptedHello = encryptedHello
	return bs.BytesRead(), nil
}

// Size 大小
func (m *MsgChangeCipherSpec) Size() int {
	return binaryutil.SizeofBytes(m.EncryptedHello)
}

// MsgId 消息Id
func (MsgChangeCipherSpec) MsgId() MsgId {
	return MsgId_ChangeCipherSpec
}

// Clone 克隆消息对象
func (m *MsgChangeCipherSpec) Clone() Msg {
	return &MsgChangeCipherSpec{
		EncryptedHello: bytes.Clone(m.EncryptedHello),
	}
}
