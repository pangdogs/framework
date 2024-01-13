package transport

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/plugins/gtp"
	"git.golaxy.org/plugins/util/binaryutil"
	"io"
	"sync/atomic"
)

// NewSequencedSynchronizer 创建有时序同步器缓存
func NewSequencedSynchronizer(sendSeq, recvSeq uint32, cap int) ISynchronizer {
	s := &SequencedSynchronizer{}
	s.Reset(sendSeq, recvSeq, cap)
	return s
}

// _SequencedFrame 时序帧
type _SequencedFrame struct {
	seq    uint32 // 序号
	offset int    // 帧数据偏移位置
	data   []byte // 帧数据
}

// SequencedSynchronizer 有时序同步器缓存，主要用于断线重连时，同步对端时序，补发消息
type SequencedSynchronizer struct {
	sendSeq uint32            // 发送消息序号
	recvSeq uint32            // 接收消息序号
	ackSeq  uint32            // 当前ack序号
	cap     int               // 缓存区容量，缓存区满时将会触发清理操作，此时断线重连有可能会失败
	cached  int               // 已缓存大小
	sent    int               // 已发送位置
	frames  []_SequencedFrame // 帧队列
}

// Reset 重置缓存
func (s *SequencedSynchronizer) Reset(sendSeq, recvSeq uint32, cap int) {
	s.sendSeq = sendSeq
	s.recvSeq = recvSeq
	s.ackSeq = sendSeq - 1
	s.cap = cap
	s.cached = 0
	s.sent = 0
	s.frames = s.frames[:]
}

// Write implements io.Writer
func (s *SequencedSynchronizer) Write(p []byte) (n int, err error) {
	// ack消息序号
	s.ack(s.getRemoteAck())

	// 缓存区满时，清理缓存
	if s.cached+len(p) > s.cap {
		s.reduce(len(p))
	}

	data := binaryutil.BytesPool.Get(len(p))
	copy(data, p)

	head := gtp.MsgHead{}
	if _, err = head.Write(data); err != nil {
		binaryutil.BytesPool.Put(data)
		return 0, err
	}

	// 填充序号
	head.Seq = s.sendSeq
	head.Ack = s.getLocalAck()

	if _, err = head.Read(data); err != nil {
		binaryutil.BytesPool.Put(data)
		return 0, err
	}

	// 写入帧队列
	s.frames = append(s.frames, _SequencedFrame{seq: head.Seq, data: data})
	s.cached += len(data)

	// 自增序号
	s.sendSeq++

	return len(data), nil
}

// WriteTo implements io.WriteTo
func (s *SequencedSynchronizer) WriteTo(w io.Writer) (int64, error) {
	if w == nil {
		return 0, fmt.Errorf("%w: w is nil", core.ErrArgs)
	}

	var wn int64

	// 读取帧队列，向输出流写入消息
	for i := s.sent; i < len(s.frames); i++ {
		frame := &s.frames[i]

		if frame.offset < len(frame.data) {
			n, err := w.Write(frame.data[frame.offset:])
			if n > 0 {
				frame.offset += n
				wn += int64(n)
			}
			if err != nil {
				return wn, err
			}
		}

		// 写入完全成功时，更新已发送位置
		s.sent++
	}

	return wn, nil
}

// Validate 验证消息包
func (s *SequencedSynchronizer) Validate(msgHead gtp.MsgHead, msgBuff []byte) error {
	// 检测消息包序号
	d := int32(msgHead.Seq - s.recvSeq)
	if d > 0 {
		return ErrUnexpectedSeq
	} else if d < 0 {
		return ErrDiscardSeq
	}
	return nil
}

// Synchronization 同步对端时序，对齐缓存序号
func (s *SequencedSynchronizer) Synchronization(remoteRecvSeq uint32) error {
	// 从时序帧中查询对端序号
	for i := len(s.frames) - 1; i >= 0; i-- {
		frame := &s.frames[i]

		d := int32(frame.seq - remoteRecvSeq)
		if d <= 0 {
			for j := i; j < len(s.frames); j++ {
				s.frames[j].offset = 0
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

	return fmt.Errorf("frame %d not found", remoteRecvSeq)
}

// Ack 确认消息序号
func (s *SequencedSynchronizer) Ack(ack uint32) error {
	// 自增接收消息序号
	atomic.AddUint32(&s.recvSeq, 1)
	// 记录ack序号
	atomic.StoreUint32(&s.ackSeq, ack)

	return nil
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
	s.sent = 0
	for i := range s.frames {
		binaryutil.BytesPool.Put(s.frames[i].data)
	}
	s.frames = nil
}

func (s *SequencedSynchronizer) getLocalAck() uint32 {
	return atomic.LoadUint32(&s.recvSeq)
}

func (s *SequencedSynchronizer) getRemoteAck() uint32 {
	return atomic.LoadUint32(&s.ackSeq)
}

func (s *SequencedSynchronizer) ack(seq uint32) {
	cached := s.cached

	for i := range s.frames {
		frame := &s.frames[i]

		cached -= len(frame.data)

		if frame.seq == seq {
			for j := 0; j <= i; j++ {
				binaryutil.BytesPool.Put(s.frames[j].data)
			}

			s.frames = append(s.frames[:0], s.frames[i+1:]...)
			s.sent = 0

			s.cached = cached
			if s.cached < 0 {
				panic("sequenced buffer cached less 0 invalid")
			}

			break
		}
	}
}

func (s *SequencedSynchronizer) reduce(size int) {
	cached := s.cached

	for i := 0; i < s.sent; i++ {
		frame := &s.frames[i]

		cached -= len(frame.data)

		size -= len(frame.data)
		if size <= 0 {
			for j := 0; j <= i; j++ {
				binaryutil.BytesPool.Put(s.frames[j].data)
			}

			s.frames = append(s.frames[:0], s.frames[i+1:]...)
			s.sent = 0

			s.cached = cached
			if s.cached < 0 {
				panic("sequenced buffer cached less 0 invalid")
			}

			break
		}
	}
}
