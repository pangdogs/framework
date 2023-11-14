package gtp

import (
	"kit.golaxy.org/plugins/util/binaryutil"
	"strings"
)

// Code 错误码
type Code int32

const (
	Code_VersionError    Code = iota + 1 // 版本错误
	Code_SessionNotFound                 // Session未找到
	Code_EncryptFailed                   // 加密失败
	Code_AuthFailed                      // 鉴权失败
	Code_ContinueFailed                  // 重连失败
	Code_Reject                          // 拒绝连接
	Code_Shutdown                        // 服务关闭
	Code_SessionDeath                    // 会话过期
	Code_Customize                       // 自定义错误码起点
)

// MsgRst 重置链路（注意：为了提高解码性能，减少内存碎片，解码string与bytes字段时均使用引用类型，引用字节池中的bytes，GC时会被归还字节池，不要直接持有此类型字段）
type MsgRst struct {
	Code    Code   // 错误码
	Message string // 错误信息
}

// Read implements io.Reader
func (m *MsgRst) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteInt32(int32(m.Code)); err != nil {
		return 0, err
	}
	if err := bs.WriteString(m.Message); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgRst) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	code, err := bs.ReadInt32()
	if err != nil {
		return 0, err
	}
	msg, err := bs.ReadStringRef()
	if err != nil {
		return 0, err
	}
	m.Code = Code(code)
	m.Message = msg
	return bs.BytesRead(), nil
}

// Size 大小
func (m *MsgRst) Size() int {
	return binaryutil.SizeofInt32() + binaryutil.SizeofString(m.Message)
}

// MsgId 消息Id
func (MsgRst) MsgId() MsgId {
	return MsgId_Rst
}

// Clone 克隆消息对象
func (m *MsgRst) Clone() Msg {
	return &MsgRst{
		Code:    m.Code,
		Message: strings.Clone(m.Message),
	}
}
