package transport

import (
	"bytes"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/plugins/gtp"
	"io"
)

// NewUnsequencedSynchronizer 创建无时序同步器缓存
func NewUnsequencedSynchronizer() ISynchronizer {
	return &UnsequencedSynchronizer{}
}

// UnsequencedSynchronizer 无时序同步器缓存
type UnsequencedSynchronizer struct {
	bytes.Buffer
}

// WriteTo implements io.WriteTo
func (s *UnsequencedSynchronizer) WriteTo(w io.Writer) (int64, error) {
	if w == nil {
		return 0, fmt.Errorf("%w: w is nil", core.ErrArgs)
	}
	return s.Buffer.WriteTo(w)
}

// Validate 验证消息包
func (s *UnsequencedSynchronizer) Validate(msgHead gtp.MsgHead, msgBuff []byte) error {
	return nil
}

// Synchronization 同步对端时序，对齐缓存序号
func (s *UnsequencedSynchronizer) Synchronization(remoteRecvSeq uint32) error {
	return errors.New("not support")
}

// Ack 确认消息序号
func (s *UnsequencedSynchronizer) Ack(ack uint32) error {
	return nil
}

// SendSeq 发送消息序号
func (s *UnsequencedSynchronizer) SendSeq() uint32 {
	return 0
}

// RecvSeq 接收消息序号
func (s *UnsequencedSynchronizer) RecvSeq() uint32 {
	return 0
}

// AckSeq 当前ack序号
func (s *UnsequencedSynchronizer) AckSeq() uint32 {
	return 0
}

// Cached 已缓存大小
func (s *UnsequencedSynchronizer) Cached() int {
	return s.Len()
}

// Clean 清理
func (s *UnsequencedSynchronizer) Clean() {
	s.Buffer.Reset()
}
