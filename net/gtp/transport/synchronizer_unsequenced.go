/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package transport

import (
	"bytes"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/framework/net/gtp"
	"io"
)

// NewUnsequencedSynchronizer 创建无时序同步器，不支持断连重连时同步时序
func NewUnsequencedSynchronizer() ISynchronizer {
	return &UnsequencedSynchronizer{}
}

// UnsequencedSynchronizer 无时序同步器，不支持断连重连时补发消息
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
func (s *UnsequencedSynchronizer) Validate(msgHead gtp.MsgHead, msgBuf []byte) error {
	return nil
}

// Synchronization 同步对端时序，对齐缓存序号
func (s *UnsequencedSynchronizer) Synchronization(remoteRecvSeq uint32) error {
	return errors.New("not support")
}

// Ack 确认消息序号
func (s *UnsequencedSynchronizer) Ack(ack uint32) {
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
