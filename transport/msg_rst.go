package transport

import "kit.golaxy.org/plugins/transport/binaryutil"

// Code 错误码
type Code = int32

const (
	Code_SessionNotFound Code = iota + 1 // Session找不到
	Code_EncryptFailed                   // 加密失败
	Code_AuthFailed                      // 鉴权失败
	Code_ContinueFailed                  // 恢复Session失败
	Code_Reject                          // 拒绝连接
	Code_ServiceShutdown                 // 服务关闭
	Code_Customize                       // 自定义错误码起点
)

// MsgRst 重置链路
type MsgRst struct {
	Code       Code   // 错误码
	Extensions []byte // 扩展内容
}

func (m *MsgRst) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	code, err := bs.ReadInt32()
	if err != nil {
		return 0, err
	}
	extensions, err := bs.ReadBytes()
	if err != nil {
		return 0, err
	}
	m.Code = code
	m.Extensions = extensions
	return bs.BytesRead(), nil
}

func (m *MsgRst) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteInt32(m.Code); err != nil {
		return 0, err
	}
	if err := bs.WriteBytes(m.Extensions); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (m *MsgRst) Size() int {
	return binaryutil.SizeofUint32() + binaryutil.SizeofBytes(m.Extensions)
}
