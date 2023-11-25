package gtp

import (
	"kit.golaxy.org/plugins/util/binaryutil"
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
		return 0, err
	}
	if err := bs.WriteUint32(m.RecvSeq); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgContinue) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	sendSeq, err := bs.ReadUint32()
	if err != nil {
		return 0, err
	}
	recvSeq, err := bs.ReadUint32()
	if err != nil {
		return 0, err
	}
	m.SendSeq = sendSeq
	m.RecvSeq = recvSeq
	return bs.BytesRead(), nil
}

// Size 大小
func (MsgContinue) Size() int {
	return binaryutil.SizeofUint32() + binaryutil.SizeofUint32() + binaryutil.SizeofUint32()
}

// MsgId 消息Id
func (MsgContinue) MsgId() MsgId {
	return MsgId_Continue
}

// Clone 克隆消息对象
func (m MsgContinue) Clone() Msg {
	return &m
}
