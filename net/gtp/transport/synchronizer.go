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
	"io"

	"git.golaxy.org/framework/net/gtp/codec"
)

var (
	// ErrSynchronizer 是消息时序同步错误的根错误。
	ErrSynchronizer = errors.New("gtp-synchronizer")
	// ErrUnexpectedSeq 表示收到的消息序号不连续。
	ErrUnexpectedSeq = fmt.Errorf("%w: unexpected sequence", ErrSynchronizer)
	// ErrDiscardSeq 表示消息序号已处理过，应丢弃重复消息。
	ErrDiscardSeq = fmt.Errorf("%w: discard sequence", ErrSynchronizer)
)

// ISynchronizer 校验消息序号、缓存待确认发送包并支持重连补发。
type ISynchronizer interface {
	io.Writer
	io.WriterTo
	codec.IValidation
	// Synchronize 根据对端已接收序号丢弃已确认缓存，并准备补发剩余消息。
	Synchronize(remoteRecvSeq uint32) error
	// Ack 确认对端已接收的发送序号。
	Ack(ack uint32)
	// SendSeq 返回当前发送序号。
	SendSeq() uint32
	// RecvSeq 返回当前接收序号。
	RecvSeq() uint32
	// AckSeq 返回对端最近确认的发送序号。
	AckSeq() uint32
	// Cap 返回发送缓存的字节容量。
	Cap() int
	// Cached 返回当前缓存的发送字节数。
	Cached() int
	// Dispose 释放缓存的池化字节缓冲区。
	Dispose()
}
