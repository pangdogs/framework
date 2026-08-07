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
	"io"
	"sync/atomic"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/utils/binaryutil"
)

// NewSequencedSynchronizer 创建从指定序号开始、缓存容量目标为 cap 字节的有序同步器。
func NewSequencedSynchronizer(sendSeq, recvSeq uint32, cap int) ISynchronizer {
	s := &SequencedSynchronizer{}
	s.init(sendSeq, recvSeq, cap)
	return s
}

const (
	queueMinSize = 16
)

// _Frame 保存一个待确认消息包及其当前发送偏移。
type _Frame struct {
	seq    uint32 // 消息序号。
	offset int    // 已发送字节偏移。
	data   []byte // 从池中取得的完整消息包。
}

// _Queue 是容量保持为二次幂的帧环形队列。
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
		exception.Panicf("%w: queue: Peek() called on empty queue", ErrSynchronizer)
	}
	return &q.buf[q.head]
}

func (q *_Queue) Index(i int) *_Frame {
	if i < 0 {
		i += q.count
	}
	if i < 0 || i >= q.count {
		exception.Panicf("%w: queue: Index() called with index out of range", ErrSynchronizer)
	}
	return &q.buf[(q.head+i)&(len(q.buf)-1)]
}

func (q *_Queue) Pop() _Frame {
	if q.count <= 0 {
		exception.Panicf("%w: queue: Pop() called on empty queue", ErrSynchronizer)
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

// SequencedSynchronizer 为消息分配序号并缓存未确认数据，以支持断线续传。
// 它由 Transceiver 的发送锁和接收锁协调，调用方不应绕过 Transceiver 并发操作。
type SequencedSynchronizer struct {
	sendSeq uint32  // 下一个发送消息序号。
	recvSeq uint32  // 下一个期望接收的消息序号。
	ackSeq  uint32  // 对端最近确认的发送序号。
	cap     int     // 缓存容量目标；超出时会淘汰已发送帧，可能导致续传失败。
	cached  int     // 当前缓存字节数。
	queue   *_Queue // 发送窗口队列。
	sent    int     // 队列中已完整写入连接的帧数。
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

// Write 为编码消息包填入序号和确认号，并复制到发送窗口。
func (s *SequencedSynchronizer) Write(p []byte) (n int, err error) {
	// 只解析头部即可校验序号并覆盖本地时序字段。
	head := gtp.MsgHead{}
	if _, err = head.Write(p); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrSynchronizer, err)
	}

	// 先淘汰对端已确认的发送帧，释放补发缓存空间。
	s.ack(s.getRemoteAck())

	// 超出容量目标时淘汰可清理的已发送帧。
	if s.cached+len(p) > s.cap {
		s.reduce(len(p))
	}

	// 出站帧携带本地发送序号和当前期望接收序号。
	head.Seq = s.sendSeq
	head.Ack = s.getLocalAck()

	// 每个帧持有独立副本，调用方可在 Write 后立即复用输入。
	data := binaryutil.BytesPool.Get(len(p))
	copy(data, p)

	if _, err = binaryutil.CopyToBuff(data, head); err != nil {
		binaryutil.BytesPool.Put(data)
		return 0, fmt.Errorf("%w: %w", ErrSynchronizer, err)
	}

	// 入队成功后推进下一发送序号。
	s.queue.Push(_Frame{seq: s.sendSeq, data: data})
	s.cached += len(data)
	s.sendSeq++

	return len(data), nil
}

// WriteTo 从上次偏移继续向 w 写出尚未完整发送的帧。
func (s *SequencedSynchronizer) WriteTo(w io.Writer) (int64, error) {
	if w == nil {
		return 0, fmt.Errorf("%w: %w: w is nil", ErrSynchronizer, core.ErrArgs)
	}

	var wn int64

	// 从 sent 指向的帧和帧内 offset 继续，支持短写后的断点补发。
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
			if frame.offset < len(frame.data) {
				exception.Panicf("%w: broken writer, wrote: %d, offset: %d, frame: %d", ErrSynchronizer, n, frame.offset, len(frame.data))
			}
		}

		// 整帧写完后才推进 sent；短写位置保存在 frame.offset。
		s.sent++
	}

	return wn, nil
}

// Validate 要求消息序号等于下一个期望接收序号。
func (s *SequencedSynchronizer) Validate(msgHead gtp.MsgHead, msgBuf []byte) error {
	// 只接受当前期望序号，旧帧丢弃，超前帧触发时序错误。
	d := int32(msgHead.Seq - s.recvSeq)
	if d > 0 {
		return ErrUnexpectedSeq
	} else if d < 0 {
		return ErrDiscardSeq
	}
	return nil
}

// Synchronize 根据对端下一个期望序号重置补发位置；所需帧已被淘汰时返回错误。
func (s *SequencedSynchronizer) Synchronize(remoteRecvSeq uint32) error {
	// 从最新帧反向寻找对端期望序号对应的补发起点。
	for i := s.queue.Length() - 1; i >= 0; i-- {
		frame := s.queue.Index(i)

		d := int32(frame.seq - remoteRecvSeq)
		if d <= 0 {
			for j := i; j < s.queue.Length(); j++ {
				s.queue.Index(j).offset = 0
			}

			s.sent = i
			s.ackSeq = frame.seq

			return nil
		}
	}

	// 对端已追上尚未分配的下一序号时无需补发。
	if s.sendSeq == remoteRecvSeq {
		return nil
	}

	return fmt.Errorf("%w: frame %d not found", ErrSynchronizer, remoteRecvSeq)
}

// Ack 在成功接收一包后推进接收序号，并记录对端确认号。
func (s *SequencedSynchronizer) Ack(ack uint32) {
	// 完整接收一帧后推进下一期望序号，并记录对端确认号。
	atomic.AddUint32(&s.recvSeq, 1)
	atomic.StoreUint32(&s.ackSeq, ack)
}

// SendSeq 返回下一个发送消息序号。
func (s *SequencedSynchronizer) SendSeq() uint32 {
	return s.sendSeq
}

// RecvSeq 返回下一个期望接收的消息序号。
func (s *SequencedSynchronizer) RecvSeq() uint32 {
	return s.recvSeq
}

// AckSeq 返回对端最近确认的发送序号。
func (s *SequencedSynchronizer) AckSeq() uint32 {
	return s.ackSeq
}

// Cap 返回发送缓存容量目标。
func (s *SequencedSynchronizer) Cap() int {
	return s.cap
}

// Cached 返回当前缓存的完整消息包字节数。
func (s *SequencedSynchronizer) Cached() int {
	return s.cached
}

// Dispose 归还全部池化帧并将序号和缓存状态清零。
func (s *SequencedSynchronizer) Dispose() {
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
	count := 0

	for i := 0; i < s.queue.Length(); i++ {
		frame := s.queue.Index(i)

		if int32(frame.seq-seq) >= 0 {
			break
		}

		cached -= len(frame.data)
		count++
	}

	if count <= 0 {
		return
	}

	for i := 0; i < count; i++ {
		binaryutil.BytesPool.Put(s.queue.Pop().data)
	}

	s.sent = max(0, s.sent-count)

	s.cached = cached
	if s.cached < 0 {
		exception.Panicf("%w: cached size underflow: %d", ErrSynchronizer, s.cached)
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
				exception.Panicf("%w: cached size underflow: %d", ErrSynchronizer, s.cached)
			}

			return
		}
	}
}
