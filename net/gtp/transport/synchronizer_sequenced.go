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
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
	"sync/atomic"
)

// NewSequencedSynchronizer 创建有时序同步器，支持缓存已发送的消息，在断连重连时同步时序并补发消息
func NewSequencedSynchronizer(sendSeq, recvSeq uint32, cap int) ISynchronizer {
	s := &SequencedSynchronizer{}
	s.init(sendSeq, recvSeq, cap)
	return s
}

const (
	queueMinSize = 16
)

// _Frame 帧
type _Frame struct {
	seq    uint32 // 序号
	offset int    // 帧数据偏移位置
	data   []byte // 帧数据
}

// _Queue 环形队列
type _Queue struct {
	buf               []_Frame
	head, tail, count int
}

func newQueue() *_Queue {
	return &_Queue{
		buf: make([]_Frame, queueMinSize),
	}
}

func (q *_Queue) Length() int {
	return q.count
}

func (q *_Queue) Push(elem _Frame) {
	if q.count == len(q.buf) {
		q.resize()
	}

	q.buf[q.tail] = elem

	q.tail = (q.tail + 1) & (len(q.buf) - 1)
	q.count++
}

func (q *_Queue) Peek() *_Frame {
	if q.count <= 0 {
		panic(fmt.Errorf("%w: queue: Peek() called on empty queue", ErrSynchronizer))
	}
	return &q.buf[q.head]
}

func (q *_Queue) Index(i int) *_Frame {
	if i < 0 {
		i += q.count
	}
	if i < 0 || i >= q.count {
		panic(fmt.Errorf("%w: queue: Index() called with index out of range", ErrSynchronizer))
	}
	return &q.buf[(q.head+i)&(len(q.buf)-1)]
}

func (q *_Queue) Pop() _Frame {
	if q.count <= 0 {
		panic(fmt.Errorf("%w: queue: Pop() called on empty queue", ErrSynchronizer))
	}

	elem := q.buf[q.head]
	q.buf[q.head] = _Frame{}

	q.head = (q.head + 1) & (len(q.buf) - 1)
	q.count--

	if len(q.buf) > queueMinSize && (q.count<<2) == len(q.buf) {
		q.resize()
	}

	return elem
}

func (q *_Queue) resize() {
	newBuf := make([]_Frame, q.count<<1)

	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}

	q.head = 0
	q.tail = q.count
	q.buf = newBuf
}

// SequencedSynchronizer 有时序同步器，支持缓存已发送的消息，在断连重连时同步时序并补发消息
type SequencedSynchronizer struct {
	sendSeq uint32  // 发送消息序号
	recvSeq uint32  // 接收消息序号
	ackSeq  uint32  // 当前ack序号
	cap     int     // 缓存区容量（字节），缓存区满时将会触发清理操作，此时断线重连有可能会失败
	cached  int     // 已缓存大小（字节）
	queue   *_Queue // 窗口队列
	sent    int     // 已发送位置
}

func (s *SequencedSynchronizer) init(sendSeq, recvSeq uint32, cap int) {
	s.sendSeq = sendSeq
	s.recvSeq = recvSeq
	s.ackSeq = sendSeq - 1
	s.cap = cap
	s.cached = 0
	s.queue = newQueue()
	s.sent = 0
}

// Write implements io.Writer
func (s *SequencedSynchronizer) Write(p []byte) (n int, err error) {
	// 读取消息头
	head := gtp.MsgHead{}
	if _, err = head.Write(p); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrSynchronizer, err)
	}

	// ack消息序号
	s.ack(s.getRemoteAck())

	// 缓存区满时，清理缓存
	if s.cached+len(p) > s.cap {
		s.reduce(len(p))
	}

	// 填充序号
	head.Seq = s.sendSeq
	head.Ack = s.getLocalAck()

	// 分配内存并拷贝数据
	data := binaryutil.BytesPool.Get(len(p))
	copy(data, p)

	if _, err = binaryutil.CopyToBuff(data, head); err != nil {
		binaryutil.BytesPool.Put(data)
		return 0, fmt.Errorf("%w: %w", ErrSynchronizer, err)
	}

	// 写入帧队列并自增序号
	s.queue.Push(_Frame{seq: s.sendSeq, data: data})
	s.cached += len(data)
	s.sendSeq++

	return len(data), nil
}

// WriteTo implements io.WriteTo
func (s *SequencedSynchronizer) WriteTo(w io.Writer) (int64, error) {
	if w == nil {
		return 0, fmt.Errorf("%w: %w: w is nil", ErrSynchronizer, core.ErrArgs)
	}

	var wn int64

	// 读取帧队列，向输出流写入消息
	for i := s.sent; i < s.queue.Length(); i++ {
		frame := s.queue.Index(i)

		if frame.offset < len(frame.data) {
			n, err := w.Write(frame.data[frame.offset:])
			if n > 0 {
				frame.offset += n
				wn += int64(n)
			}
			if err != nil {
				return wn, fmt.Errorf("%w: %w", ErrSynchronizer, err)
			}
		}

		// 写入完全成功时，更新已发送位置
		s.sent++
	}

	return wn, nil
}

// Validate 验证消息包
func (s *SequencedSynchronizer) Validate(msgHead gtp.MsgHead, msgBuf []byte) error {
	// 检测消息包序号
	d := int32(msgHead.Seq - s.recvSeq)
	if d > 0 {
		return ErrUnexpectedSeq
	} else if d < 0 {
		return ErrDiscardSeq
	}
	return nil
}

// Synchronize 同步对端时序，对齐缓存序号
func (s *SequencedSynchronizer) Synchronize(remoteRecvSeq uint32) error {
	// 从时序帧中查询对端序号
	for i := s.queue.Length() - 1; i >= 0; i-- {
		frame := s.queue.Index(i)

		d := int32(frame.seq - remoteRecvSeq)
		if d <= 0 {
			for j := i; j < s.queue.Length(); j++ {
				s.queue.Index(j).offset = 0
			}

			s.sent = i
			s.ackSeq = frame.seq - 1

			return nil
		}
	}

	// 发送序号与对端接收序号相同
	if s.sendSeq == remoteRecvSeq {
		return nil
	}

	return fmt.Errorf("%w: frame %d not found", ErrSynchronizer, remoteRecvSeq)
}

// Ack 确认消息序号
func (s *SequencedSynchronizer) Ack(ack uint32) {
	// 自增接收消息序号
	atomic.AddUint32(&s.recvSeq, 1)
	// 记录ack序号
	atomic.StoreUint32(&s.ackSeq, ack)
}

// SendSeq 发送消息序号
func (s *SequencedSynchronizer) SendSeq() uint32 {
	return s.sendSeq
}

// RecvSeq 接收消息序号
func (s *SequencedSynchronizer) RecvSeq() uint32 {
	return s.recvSeq
}

// AckSeq 当前ack序号
func (s *SequencedSynchronizer) AckSeq() uint32 {
	return s.ackSeq
}

// Cap 缓存区容量
func (s *SequencedSynchronizer) Cap() int {
	return s.cap
}

// Cached 已缓存大小
func (s *SequencedSynchronizer) Cached() int {
	return s.cached
}

// Clean 清理
func (s *SequencedSynchronizer) Clean() {
	s.sendSeq = 0
	s.recvSeq = 0
	s.ackSeq = 0
	s.cap = 0
	s.cached = 0
	for i := 0; i < s.queue.Length(); i++ {
		binaryutil.BytesPool.Put(s.queue.Index(i).data)
	}
	s.queue = newQueue()
	s.sent = 0
}

func (s *SequencedSynchronizer) getLocalAck() uint32 {
	return atomic.LoadUint32(&s.recvSeq)
}

func (s *SequencedSynchronizer) getRemoteAck() uint32 {
	return atomic.LoadUint32(&s.ackSeq)
}

func (s *SequencedSynchronizer) ack(seq uint32) {
	cached := s.cached

	for i := 0; i < s.queue.Length(); i++ {
		frame := s.queue.Index(i)

		cached -= len(frame.data)

		if frame.seq == seq {
			for j := 0; j <= i; j++ {
				binaryutil.BytesPool.Put(s.queue.Pop().data)
			}

			s.sent = 0

			s.cached = cached
			if s.cached < 0 {
				exception.Panicf("%w: sequenced buffer cached less 0 invalid", ErrSynchronizer)
			}

			return
		}
	}
}

func (s *SequencedSynchronizer) reduce(size int) {
	cached := s.cached

	for i := 0; i < s.sent; i++ {
		frame := s.queue.Index(i)

		cached -= len(frame.data)

		size -= len(frame.data)
		if size <= 0 {
			for j := 0; j <= i; j++ {
				binaryutil.BytesPool.Put(s.queue.Pop().data)
			}

			s.sent = 0

			s.cached = cached
			if s.cached < 0 {
				exception.Panicf("%w: sequenced buffer cached less 0 invalid", ErrSynchronizer)
			}

			return
		}
	}
}
