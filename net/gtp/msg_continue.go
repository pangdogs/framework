package gtp

import (
	"git.golaxy.org/framework/utils/binaryutil"
)

// MsgContinue 重连
type MsgContinue struct {
	SendSeq uint32 // 客户端请求消息序号
	RecvSeq uint32 // 客户端响应消息序号
}

// Read implements io.Reader
func (m MsgContinue) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint32(m.SendSeq); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint32(m.RecvSeq); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgContinue) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	var err error

	m.SendSeq, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.RecvSeq, err = bs.ReadUint32()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (MsgContinue) Size() int {
	return binaryutil.SizeofUint32() + binaryutil.SizeofUint32()
}

// MsgId 消息Id
func (MsgContinue) MsgId() MsgId {
	return MsgId_Continue
}

// Clone 克隆消息对象
func (m MsgContinue) Clone() MsgReader {
	return m
}
