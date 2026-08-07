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
	"fmt"
	"io"

	"git.golaxy.org/core"
	"git.golaxy.org/framework/net/gtp"
)

// NewUnsequencedSynchronizer 创建不校验序号、也不支持断线补发的同步器。
func NewUnsequencedSynchronizer() ISynchronizer {
	return &UnsequencedSynchronizer{}
}

// UnsequencedSynchronizer 只暂存待写字节，不维护发送窗口或消息序号。
type UnsequencedSynchronizer struct {
	bytes.Buffer
}

// WriteTo 将全部暂存数据写入 w 并从缓冲区移除。
func (s *UnsequencedSynchronizer) WriteTo(w io.Writer) (int64, error) {
	if w == nil {
		return 0, fmt.Errorf("%w: %w: w is nil", ErrSynchronizer, core.ErrArgs)
	}
	return s.Buffer.WriteTo(w)
}

// Validate 接受任意消息序号。
func (s *UnsequencedSynchronizer) Validate(msgHead gtp.MsgHead, msgBuf []byte) error {
	return nil
}

// Synchronize 始终返回不支持续传的错误。
func (s *UnsequencedSynchronizer) Synchronize(remoteRecvSeq uint32) error {
	return fmt.Errorf("%w: not supported", ErrSynchronizer)
}

// Ack 不维护确认状态。
func (s *UnsequencedSynchronizer) Ack(ack uint32) {
}

// SendSeq 始终返回零。
func (s *UnsequencedSynchronizer) SendSeq() uint32 {
	return 0
}

// RecvSeq 始终返回零。
func (s *UnsequencedSynchronizer) RecvSeq() uint32 {
	return 0
}

// AckSeq 始终返回零。
func (s *UnsequencedSynchronizer) AckSeq() uint32 {
	return 0
}

// Cached 返回当前暂存字节数。
func (s *UnsequencedSynchronizer) Cached() int {
	return s.Len()
}

// Dispose 清空暂存字节。
func (s *UnsequencedSynchronizer) Dispose() {
	s.Buffer.Reset()
}
