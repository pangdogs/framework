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
	"errors"
	"fmt"
	"git.golaxy.org/framework/net/gtp/codec"
	"io"
)

var (
	ErrSynchronizer  = errors.New("gtp-synchronizer")                         // 同步器错误
	ErrUnexpectedSeq = fmt.Errorf("%w: unexpected sequence", ErrSynchronizer) // 收到非预期的消息序号，表示序号不连续
	ErrDiscardSeq    = fmt.Errorf("%w: discard sequence", ErrSynchronizer)    // 收到已过期的消息序号，表示次消息已收到过
)

// ISynchronizer 同步器
type ISynchronizer interface {
	io.Writer
	io.WriterTo
	codec.IValidation
	// Synchronize 同步对端时序，对齐缓存序号
	Synchronize(remoteRecvSeq uint32) error
	// Ack 确认消息序号
	Ack(ack uint32)
	// SendSeq 发送消息序号
	SendSeq() uint32
	// RecvSeq 接收消息序号
	RecvSeq() uint32
	// AckSeq 当前ack序号
	AckSeq() uint32
	// Cap 缓存区容量
	Cap() int
	// Cached 已缓存大小
	Cached() int
	// Clean 清理
	Clean()
}
