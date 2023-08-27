package gtp

import (
	"bytes"
	"kit.golaxy.org/plugins/gtp/binaryutil"
	"strings"
)

// MsgAuth 鉴权（注意：为了提高解码性能，减少内存碎片，解码string与bytes字段时均使用引用类型，引用字节池中的bytes，GC时会被归还字节池，不要直接持有此类型字段）
type MsgAuth struct {
	Token      string // 令牌
	Extensions []byte // 扩展内容
}

// Read implements io.Reader
func (m *MsgAuth) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteString(m.Token); err != nil {
		return 0, err
	}
	if err := bs.WriteBytes(m.Extensions); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgAuth) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	token, err := bs.ReadStringRef()
	if err != nil {
		return 0, err
	}
	extensions, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	m.Token = token
	m.Extensions = extensions
	return bs.BytesRead(), nil
}

// Size 消息大小
func (m *MsgAuth) Size() int {
	return binaryutil.SizeofString(m.Token) + binaryutil.SizeofBytes(m.Extensions)
}

// MsgId 消息Id
func (MsgAuth) MsgId() MsgId {
	return MsgId_Auth
}

// Clone 克隆消息对象
func (m *MsgAuth) Clone() Msg {
	return &MsgAuth{
		Token:      strings.Clone(m.Token),
		Extensions: bytes.Clone(m.Extensions),
	}
}
