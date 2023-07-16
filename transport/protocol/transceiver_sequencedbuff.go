package protocol

import (
	"io"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"sync"
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
	Cap     int               // 缓存区容量，缓存区满时将会触发清理操作，此时断线重连有可能会失败
	cached  int               // 已缓存大小
	sent    int               // 已发送位置
	frames  []_SequencedFrame // 帧队列
	mutex   sync.Mutex        // 锁
}

// Write implements io.Writer
func (s *SequencedBuff) Write(p []byte) (n int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var wn int64

	// 读取帧队列，向输出流写入消息
	for i := s.sent; i < len(s.frames); i++ {
		msg := &s.frames[i]

		if msg.Offset < len(msg.Data) {
			n, err := w.Write(msg.Data[msg.Offset:])
			if n > 0 {
				msg.Offset += n
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
		msg := &s.frames[i]

		if !msg.WaitAck && msg.Offset >= len(msg.Data) {
			s.frames = append(s.frames[:i], s.frames[i+1:]...)

			s.cached -= len(msg.Data)
			if s.cached < 0 {
				s.cached = 0
			}

			s.sent--
			if s.sent < 0 {
				s.sent = 0
			}

			codec.BytesPool.Put(msg.Data)
		}
	}

	return wn, nil
}

// Validation 验证消息包
func (s *SequencedBuff) Validation(mp transport.MsgPacket) error {
	if !mp.Head.Flags.Is(transport.Flag_Sequenced) {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检测消息包序号
	d := mp.Head.Seq - s.RecvSeq
	if d > 0 {
		return ErrUnexpectedSeq
	} else if d < 0 {
		return ErrDiscardSeq
	}

	s.RecvSeq++

	s.ack(mp.Head.Ack)

	return nil
}

func (s *SequencedBuff) getSeq() uint32 {
	return s.SendSeq
}

func (s *SequencedBuff) getAck() uint32 {
	return s.RecvSeq
}

func (s *SequencedBuff) ack(seq uint32) {
	cached := s.cached

	for i := range s.frames {
		msg := &s.frames[i]

		cached -= len(msg.Data)

		if msg.Seq == seq {
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
		msg := &s.frames[i]

		cached -= len(msg.Data)

		size -= len(msg.Data)
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
