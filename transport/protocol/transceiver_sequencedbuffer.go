package protocol

import (
	"fmt"
	"io"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"sync/atomic"
)

// Buffer 缓存
type Buffer interface {
	io.Writer
	io.WriterTo
	// Synchronization 同步对端时序，对齐缓存
	Synchronization(remoteRecvSeq uint32) error
	// Validation 验证消息头
	Validation(msgHead transport.MsgHead) error
	// Ack 确认消息序号
	Ack(ack uint32) error
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

// _SequencedFrame 时序帧
type _SequencedFrame struct {
	Seq    uint32 // 序号
	Offset int    // 帧数据偏移位置
	Data   []byte // 帧数据
}

// SequencedBuffer 时序缓存，主要用于断线重连时，同步对端时序，补发消息
type SequencedBuffer struct {
	sendSeq uint32            // 发送消息序号
	recvSeq uint32            // 接收消息序号
	ackSeq  uint32            // 当前ack序号
	cap     int               // 缓存区容量，缓存区满时将会触发清理操作，此时断线重连有可能会失败
	cached  int               // 已缓存大小
	sent    int               // 已发送位置
	frames  []_SequencedFrame // 帧队列
}

// Reset 重置缓存
func (s *SequencedBuffer) Reset(sendSeq, recvSeq uint32, cap int) {
	s.sendSeq = sendSeq
	s.recvSeq = recvSeq
	s.ackSeq = sendSeq - 1
	s.cap = cap
	s.cached = 0
	s.sent = 0
	s.frames = s.frames[:]
}

// Write implements io.Writer
func (s *SequencedBuffer) Write(p []byte) (n int, err error) {
	// ack消息序号
	s.ack(s.getRemoteAck())

	// 缓存区满时，清理缓存
	if s.cached+len(p) > s.cap {
		s.reduce(len(p))
	}

	data := codec.BytesPool.Get(len(p))
	copy(data, p)

	head := transport.MsgHead{}
	if _, err = head.Write(data); err != nil {
		codec.BytesPool.Put(data)
		return 0, err
	}

	// 填充序号
	head.Seq = s.sendSeq
	head.Ack = s.getLocalAck()

	if _, err = head.Read(data); err != nil {
		codec.BytesPool.Put(data)
		return 0, err
	}

	// 写入帧队列
	s.frames = append(s.frames, _SequencedFrame{Seq: head.Seq, Data: data})
	s.cached += len(data)

	// 自增序号
	s.sendSeq++

	return len(data), nil
}

// WriteTo implements io.WriteTo
func (s *SequencedBuffer) WriteTo(w io.Writer) (int64, error) {
	var wn int64

	// 读取帧队列，向输出流写入消息
	for i := s.sent; i < len(s.frames); i++ {
		frame := &s.frames[i]

		if frame.Offset < len(frame.Data) {
			n, err := w.Write(frame.Data[frame.Offset:])
			if n > 0 {
				frame.Offset += n
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

// Synchronization 同步对端时序，对齐缓存
func (s *SequencedBuffer) Synchronization(remoteRecvSeq uint32) error {
	// 序号已对齐
	if s.sendSeq == remoteRecvSeq {
		return nil
	}

	// 调整序号
	for i := len(s.frames) - 1; i >= 0; i-- {
		frame := &s.frames[i]

		if frame.Seq == remoteRecvSeq {
			for j := i; j < len(s.frames); j++ {
				s.frames[j].Offset = 0
			}

			s.sent = i
			s.sendSeq = frame.Seq
			s.ackSeq = s.sendSeq - 1

			return nil
		}
	}

	return fmt.Errorf("frame %d not found", remoteRecvSeq)
}

// Validation 验证消息头
func (s *SequencedBuffer) Validation(msgHead transport.MsgHead) error {
	// 检测消息包序号
	d := int32(msgHead.Seq - s.recvSeq)
	if d > 0 {
		return ErrUnexpectedSeq
	} else if d < 0 {
		return ErrDiscardSeq
	}
	return nil
}

// Ack 确认消息序号
func (s *SequencedBuffer) Ack(ack uint32) error {
	// 自增接收消息序号
	atomic.AddUint32(&s.recvSeq, 1)
	// 记录ack序号
	atomic.StoreUint32(&s.ackSeq, ack)

	return nil
}

// SendSeq 发送消息序号
func (s *SequencedBuffer) SendSeq() uint32 {
	return s.sendSeq
}

// RecvSeq 接收消息序号
func (s *SequencedBuffer) RecvSeq() uint32 {
	return s.recvSeq
}

// AckSeq 当前ack序号
func (s *SequencedBuffer) AckSeq() uint32 {
	return s.ackSeq
}

// Cap 缓存区容量
func (s *SequencedBuffer) Cap() int {
	return s.cap
}

// Cached 已缓存大小
func (s *SequencedBuffer) Cached() int {
	return s.cached
}

// Clean 清理
func (s *SequencedBuffer) Clean() {
	s.sendSeq = 0
	s.recvSeq = 0
	s.ackSeq = 0
	s.cap = 0
	s.cached = 0
	s.sent = 0
	for i := range s.frames {
		codec.BytesPool.Put(s.frames[i].Data)
	}
	s.frames = nil
}

func (s *SequencedBuffer) getLocalAck() uint32 {
	return atomic.LoadUint32(&s.recvSeq)
}

func (s *SequencedBuffer) getRemoteAck() uint32 {
	return atomic.LoadUint32(&s.ackSeq)
}

func (s *SequencedBuffer) ack(seq uint32) {
	cached := s.cached

	for i := range s.frames {
		frame := &s.frames[i]

		cached -= len(frame.Data)

		if frame.Seq == seq {
			for j := 0; j <= i; j++ {
				codec.BytesPool.Put(s.frames[j].Data)
			}

			s.frames = append(s.frames[:0], s.frames[i+1:]...)
			s.sent = 0

			s.cached = cached
			if s.cached < 0 {
				panic(fmt.Errorf("sequenced buffer cached less 0 invalid"))
			}

			break
		}
	}
}

func (s *SequencedBuffer) reduce(size int) {
	cached := s.cached

	for i := 0; i < s.sent; i++ {
		frame := &s.frames[i]

		cached -= len(frame.Data)

		size -= len(frame.Data)
		if size <= 0 {
			for j := 0; j <= i; j++ {
				codec.BytesPool.Put(s.frames[j].Data)
			}

			s.frames = append(s.frames[:0], s.frames[i+1:]...)
			s.sent = 0

			s.cached = cached
			if s.cached < 0 {
				panic(fmt.Errorf("sequenced buffer cached less 0 invalid"))
			}

			break
		}
	}
}
