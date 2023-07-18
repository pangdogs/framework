package transport

import "kit.golaxy.org/plugins/transport/binaryutil"

// Code 错误码
type Code int32

const (
	Code_VersionError     Code = iota + 1 // 版本错误
	Code_SessionNotFound                  // Session找不到
	Code_EncryptFailed                    // 加密失败
	Code_AuthFailed                       // 鉴权失败
	Code_ContinueFailed                   // 重连失败
	Code_Reject                           // 拒绝连接
	Code_Shutdown                         // 服务关闭
	Code_LoginFromAnother                 // 其他地点登录
	Code_Customize                        // 自定义错误码起点
)

// MsgRst 重置链路
type MsgRst struct {
	Code    Code   // 错误码
	Message string // 错误信息
}

func (m *MsgRst) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteInt32(int32(m.Code)); err != nil {
		return 0, err
	}
	if err := bs.WriteString(m.Message); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (m *MsgRst) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
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

func (m *MsgRst) Size() int {
	return binaryutil.SizeofInt32() + binaryutil.SizeofString(m.Message)
}

func (MsgRst) MsgId() MsgId {
	return MsgId_Rst
}
