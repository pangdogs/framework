package protocol

import (
	"errors"
	"fmt"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"net"
)

// Event 消息事件
type Event[T transport.Msg] struct {
	Flags transport.Flags // 标志位
	Seq   uint32          // 消息序号
	Msg   T               // 消息
}

// UnpackEvent 解包消息事件
func UnpackEvent[T transport.Msg](e Event[transport.Msg]) Event[T] {
	return Event[T]{
		Flags: e.Flags,
		Seq:   e.Seq,
		Msg:   e.Msg.(T),
	}
}

// PackEvent 打包消息事件
func PackEvent[T transport.Msg](e Event[T]) Event[transport.Msg] {
	return Event[transport.Msg]{
		Flags: e.Flags,
		Seq:   e.Seq,
		Msg:   e.Msg,
	}
}

// RstError Rst错误提示
type RstError struct {
	Code    transport.Code // 错误码
	Message string         // 错误信息
}

// Error 错误信息
func (e *RstError) Error() string {
	return fmt.Sprintf("(%d)%s", e.Code, e.Message)
}

// EventToRstErr Rst错误消息事件转换为错误提示
func EventToRstErr(e Event[*transport.MsgRst]) *RstError {
	return &RstError{
		Code:    e.Msg.Code,
		Message: e.Msg.Message,
	}
}

// RstErrToEvent Rst错误提示转换为消息事件
func RstErrToEvent(err *RstError) Event[*transport.MsgRst] {
	return Event[*transport.MsgRst]{
		Msg: &transport.MsgRst{
			Code:    err.Code,
			Message: err.Message,
		},
	}
}

// Transceiver 消息事件收发器
type Transceiver struct {
	Conn             net.Conn       // 网络连接
	Encoder          codec.IEncoder // 消息包编码器
	Decoder          codec.IDecoder // 消息包解码器
	SendSeq, RecvSeq uint32         // 请求响应序号
}

// Send 发送消息事件
func (t *Transceiver) Send(e Event[transport.Msg], sequenced bool) error {
	if t.Conn == nil {
		return errors.New("conn is nil")
	}

	if t.Encoder == nil {
		return errors.New("encoder is nil")
	}

	if err := t.Encoder.Stuff(e.Flags.Setd(transport.Flag_Sequenced, sequenced), t.SendSeq, e.Msg); err != nil {
		return fmt.Errorf("stuff event msg failed, %w", err)
	}

	if sequenced {
		t.SendSeq++
	}

	if _, err := t.Encoder.WriteTo(t.Conn); err != nil {
		return fmt.Errorf("send msg-packet failed, %w", err)
	}

	return nil
}

// SendRst 发送Rst消息事件
func (t *Transceiver) SendRst(err error) error {
	if t.Conn == nil {
		return errors.New("conn is nil")
	}

	if t.Encoder == nil {
		return errors.New("encoder is nil")
	}

	msg := &transport.MsgRst{}

	rstErr, ok := err.(*RstError)
	if ok {
		msg.Code = rstErr.Code
		msg.Message = rstErr.Message
	} else {
		msg.Code = transport.Code_Reject
		if err != nil {
			msg.Message = err.Error()
		}
	}

	if err := t.Encoder.Stuff(transport.Flags_None(), 0, msg); err != nil {
		return err
	}

	if _, err := t.Encoder.WriteTo(t.Conn); err != nil {
		return err
	}

	return nil
}

// Recv 接收单个消息事件
func (t *Transceiver) Recv() (Event[transport.Msg], error) {
	if t.Conn == nil {
		return Event[transport.Msg]{}, errors.New("conn is nil")
	}

	if t.Decoder == nil {
		return Event[transport.Msg]{}, errors.New("decoder is nil")
	}

	for {
		var recvMP transport.MsgPacket

		if err := t.Decoder.Fetch(func(mp transport.MsgPacket) { recvMP = mp }); err != nil {
			if !errors.Is(err, codec.ErrEmptyBuffer) {
				return Event[transport.Msg]{}, fmt.Errorf("fetch recv msg-packet failed, %w", err)
			}
		} else {
			if recvMP.Head.Flags.Is(transport.Flag_Sequenced) {
				t.RecvSeq++
			}
			return Event[transport.Msg]{
				Flags: recvMP.Head.Flags,
				Seq:   recvMP.Head.Seq,
				Msg:   recvMP.Msg,
			}, nil
		}

		if _, err := t.Decoder.ReadFrom(t.Conn); err != nil {
			return Event[transport.Msg]{}, fmt.Errorf("recv msg-packet failed, %w", err)
		}
	}
}

// MultiRecv 接收多个消息事件
func (t *Transceiver) MultiRecv(fun func(Event[transport.Msg]) bool) error {
	if fun == nil {
		return errors.New("fun is nil")
	}

	if t.Conn == nil {
		return errors.New("conn is nil")
	}

	if t.Decoder == nil {
		return errors.New("decoder is nil")
	}

	for {
		var b bool

		err := t.Decoder.MultiFetch(func(mp transport.MsgPacket) bool {
			if mp.Head.Flags.Is(transport.Flag_Sequenced) {
				t.RecvSeq++
			}
			b = fun(Event[transport.Msg]{
				Flags: mp.Head.Flags,
				Seq:   mp.Head.Seq,
				Msg:   mp.Msg,
			})
			return b
		})
		if err != nil {
			if !errors.Is(err, codec.ErrEmptyBuffer) {
				return fmt.Errorf("fetch recv msg-packet failed, %w", err)
			}
		}
		if !b {
			return err
		}

		if _, err := t.Decoder.ReadFrom(t.Conn); err != nil {
			return fmt.Errorf("recv msg-packet failed, %w", err)
		}
	}
}
