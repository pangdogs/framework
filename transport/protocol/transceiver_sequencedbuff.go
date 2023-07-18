package protocol

import (
	"io"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"sync/atomic"
)

// _SequencedFrame 时序帧
type _SequencedFrame struct {
	Seq     uint32 // 序号
	Offset  int    // 帧数据偏移位置
	WaitAck bool   // 等待ACK后删除
	Data    []byte // 帧数据
}

// SequencedBuff 时序缓存
type SequencedBuff struct {
	SendSeq uint32            // 发送消息序号
	RecvSeq uint32            // 接收消息序号
	AckSeq  uint32            // 当前ack序号
	Cap     int               // 缓存区容量，缓存区满时将会触发清理操作，此时断线重连有可能会失败
	cached  int               // 已缓存大小
	sent    int               // 已发送位置
	frames  []_SequencedFrame // 帧队列
}

// Reset 重置缓存
func (s *SequencedBuff) Reset(sendSeq, recvSeq uint32) {
	s.SendSeq = sendSeq
	s.RecvSeq = recvSeq
	s.AckSeq = sendSeq - 1
	s.cached = 0
	s.sent = 0
	s.frames = s.frames[:]
}

// Synchronization 同步对端时序
func (s *SequencedBuff) Synchronization(remoteRecvSeq uint32) bool {
	// 序号已对齐
	if s.SendSeq == remoteRecvSeq {
		return true
	}

	// 调整序号
	for i := s.sent - 1; i >= 0; i-- {
		frame := &s.frames[i]

		if frame.WaitAck && frame.Seq == remoteRecvSeq {
			s.sent = i

			for j := s.sent; j < len(s.frames); j++ {
				s.frames[j].Offset = 0
			}

			s.SendSeq = frame.Seq
			s.AckSeq = s.SendSeq - 1

			return true
		}
	}
	return false
}

// Write implements io.Writer
func (s *SequencedBuff) Write(p []byte) (n int, err error) {
	// ack消息序号
	s.ack(s.getRemoteAck())

	// 缓存区满时，清理缓存
	if s.cached+len(p) > s.Cap {
		s.reduce(len(p))
	}

	data := codec.BytesPool.Get(len(p))
	copy(data, p)

	head := transport.MsgHead{}
	if _, err = head.Write(data); err != nil {
		codec.BytesPool.Put(data)
		return 0, err
	}

	// 有时序消息填充序号
	if head.Flags.Is(transport.Flag_Sequenced) {
		head.Seq = s.getSeq()
		head.Ack = s.getAck()
	} else {
		head.Seq = 0
		head.Ack = 0
	}

	if _, err = head.Read(data); err != nil {
		codec.BytesPool.Put(data)
		return 0, err
	}

	// 写入帧队列
	s.frames = append(s.frames, _SequencedFrame{Seq: head.Seq, WaitAck: head.Flags.Is(transport.Flag_Sequenced), Data: data})
	s.cached += len(data)

	// 有时序消息自增序号
	if head.Flags.Is(transport.Flag_Sequenced) {
		s.SendSeq++
	}

	return len(data), nil
}

// WriteTo implements io.WriteTo
func (s *SequencedBuff) WriteTo(w io.Writer) (int64, error) {
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

	// 删除帧队列中已发送的无时序消息
	for i := s.sent - 1; i >= 0; i-- {
		frame := &s.frames[i]

		if !frame.WaitAck && frame.Offset >= len(frame.Data) {
			s.frames = append(s.frames[:i], s.frames[i+1:]...)

			s.cached -= len(frame.Data)
			if s.cached < 0 {
				s.cached = 0
			}

			s.sent--
			if s.sent < 0 {
				s.sent = 0
			}

			codec.BytesPool.Put(frame.Data)
		}
	}

	return wn, nil
}

// Validation 验证消息包
func (s *SequencedBuff) Validation(mp transport.MsgPacket) error {
	if !mp.Head.Flags.Is(transport.Flag_Sequenced) {
		return nil
	}

	// 检测消息包序号
	d := int32(mp.Head.Seq - s.RecvSeq)
	if d > 0 {
		return ErrUnexpectedSeq
	} else if d < 0 {
		return ErrDiscardSeq
	}

	// 自增接收消息序号
	atomic.AddUint32(&s.RecvSeq, 1)

	// 记录ack序号
	atomic.StoreUint32(&s.AckSeq, mp.Head.Ack)

	return nil
}

func (s *SequencedBuff) getSeq() uint32 {
	return s.SendSeq
}

func (s *SequencedBuff) getAck() uint32 {
	return atomic.LoadUint32(&s.RecvSeq)
}

func (s *SequencedBuff) getRemoteAck() uint32 {
	return atomic.LoadUint32(&s.AckSeq)
}

func (s *SequencedBuff) ack(seq uint32) {
	cached := s.cached

	for i := range s.frames {
		frame := &s.frames[i]

		cached -= len(frame.Data)

		if frame.Seq == seq {
			for j := 0; j <= i; j++ {
				codec.BytesPool.Put(s.frames[j].Data)
			}

			s.frames = append(s.frames[:0], s.frames[i+1:]...)

			s.cached = cached
			if s.cached < 0 {
				s.cached = 0
			}

			s.sent -= i + 1
			if s.sent < 0 {
				s.sent = 0
			}

			break
		}
	}
}

func (s *SequencedBuff) reduce(size int) {
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

			s.cached = cached
			if s.cached < 0 {
				s.cached = 0
			}

			s.sent -= i + 1
			if s.sent < 0 {
				s.sent = 0
			}

			break
		}
	}
}
