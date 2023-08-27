package transport

import (
	"bytes"
	"errors"
	"kit.golaxy.org/plugins/gtp"
)

// UnsequencedBuffer 非时序缓存
type UnsequencedBuffer struct {
	bytes.Buffer
}

// Synchronization 同步对端时序，对齐缓存序号
func (s *UnsequencedBuffer) Synchronization(remoteRecvSeq uint32) error {
	return errors.New("not support")
}

// Validation 验证消息头
func (s *UnsequencedBuffer) Validation(msgHead gtp.MsgHead) error {
	return nil
}

// Ack 确认消息序号
func (s *UnsequencedBuffer) Ack(ack uint32) error {
	return nil
}

// SendSeq 发送消息序号
func (s *UnsequencedBuffer) SendSeq() uint32 {
	return 0
}

// RecvSeq 接收消息序号
func (s *UnsequencedBuffer) RecvSeq() uint32 {
	return 0
}

// AckSeq 当前ack序号
func (s *UnsequencedBuffer) AckSeq() uint32 {
	return 0
}

// Cached 已缓存大小
func (s *UnsequencedBuffer) Cached() int {
	return s.Len()
}

// Clean 清理
func (s *UnsequencedBuffer) Clean() {
	s.Buffer.Reset()
}
