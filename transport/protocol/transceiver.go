package protocol

import (
	"errors"
	"fmt"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"net"
	"sync"
	"time"
)

var (
	ErrUnexpectedSeq = errors.New("unexpected sequence") // 收到非预期的消息序号，表示序号不连续
	ErrDiscardSeq    = errors.New("discard sequence")    // 收到已过期的消息序号，表示次消息已收到过
)

// Transceiver 消息事件收发器
type Transceiver struct {
	Conn                 net.Conn       // 网络连接
	Encoder              codec.IEncoder // 消息包编码器
	Decoder              codec.IDecoder // 消息包解码器
	Timeout              time.Duration  // 网络io超时时间
	SequencedBuff        SequencedBuff  // 时序缓冲
	sendMutex, recvMutex sync.Mutex     // 发送与接收消息锁
}

// Send 发送消息事件
func (t *Transceiver) Send(e Event[transport.Msg]) error {
	if t.Conn == nil {
		return errors.New("conn is nil")
	}

	if t.Encoder == nil {
		return errors.New("encoder is nil")
	}

	t.sendMutex.Lock()
	defer t.sendMutex.Unlock()

	// 编码消息
	if err := t.Encoder.StuffTo(&t.SequencedBuff, e.Flags, e.Msg); err != nil {
		return fmt.Errorf("stuff msg failed, %w", err)
	}

	if t.Timeout > 0 {
		if err := t.Conn.SetWriteDeadline(time.Now().Add(t.Timeout)); err != nil {
			return fmt.Errorf("set conn send timeout failed, %w", err)
		}
	}

	// 数据写入链路
	if _, err := t.SequencedBuff.WriteTo(t.Conn); err != nil {
		return fmt.Errorf("send msg-packet failed, %w", err)
	}

	return nil
}

// SendRst 发送Rst消息事件
func (t *Transceiver) SendRst(err error) error {
	// 包装错误信息
	rstErr, ok := err.(*RstError)
	if !ok {
		rstErr = &RstError{Code: transport.Code_Reject}
		if err != nil {
			rstErr.Message = err.Error()
		}
	}
	return t.Send(PackEvent(RstErrToEvent(rstErr)))
}

// Recv 接收消息事件
func (t *Transceiver) Recv() (Event[transport.Msg], error) {
	if t.Conn == nil {
		return Event[transport.Msg]{}, errors.New("conn is nil")
	}

	if t.Decoder == nil {
		return Event[transport.Msg]{}, errors.New("decoder is nil")
	}

	t.recvMutex.Lock()
	defer t.recvMutex.Unlock()

	for {
		// 解码消息
		mp, err := t.Decoder.Fetch()
		if err != nil {
			if !errors.Is(err, codec.ErrEmptyBuffer) {
				return Event[transport.Msg]{}, fmt.Errorf("fetch msg-packet failed, %w", err)
			}
		} else {
			return Event[transport.Msg]{
				Flags: mp.Head.Flags,
				Seq:   mp.Head.Seq,
				Ack:   mp.Head.Ack,
				Msg:   mp.Msg,
			}, t.SequencedBuff.Validation(mp)
		}

		if t.Timeout > 0 {
			if err := t.Conn.SetReadDeadline(time.Now().Add(t.Timeout)); err != nil {
				return Event[transport.Msg]{}, fmt.Errorf("set conn recv timeout failed, %w", err)
			}
		}

		// 从链路读取消息
		if _, err := t.Decoder.ReadFrom(t.Conn); err != nil {
			return Event[transport.Msg]{}, fmt.Errorf("recv msg-packet failed, %w", err)
		}
	}
}
