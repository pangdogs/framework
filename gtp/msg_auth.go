package gtp

import (
	"bytes"
	"kit.golaxy.org/plugins/util/binaryutil"
	"strings"
)

// MsgAuth 鉴权（注意：为了提高解码性能，减少内存碎片，解码string与bytes字段时均使用引用类型，引用字节池中的bytes，GC时会被归还字节池，不要直接持有此类型字段）
type MsgAuth struct {
	UserId     string // 用户Id
	Token      string // 令牌
	Extensions []byte // 扩展内容
}

// Read implements io.Reader
func (m MsgAuth) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteString(m.UserId); err != nil {
		return 0, err
	}
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
	bs := binaryutil.NewBigEndianStream(p)
	userId, err := bs.ReadStringRef()
	if err != nil {
		return 0, err
	}
	token, err := bs.ReadStringRef()
	if err != nil {
		return 0, err
	}
	extensions, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	m.UserId = userId
	m.Token = token
	m.Extensions = extensions
	return bs.BytesRead(), nil
}

// Size 大小
func (m MsgAuth) Size() int {
	return binaryutil.SizeofString(m.UserId) + binaryutil.SizeofString(m.Token) + binaryutil.SizeofBytes(m.Extensions)
}

// MsgId 消息Id
func (MsgAuth) MsgId() MsgId {
	return MsgId_Auth
}

// Clone 克隆消息对象
func (m *MsgAuth) Clone() Msg {
	return &MsgAuth{
		UserId:     strings.Clone(m.UserId),
		Token:      strings.Clone(m.Token),
		Extensions: bytes.Clone(m.Extensions),
	}
}
