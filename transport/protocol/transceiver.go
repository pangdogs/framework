package protocol

import (
	"errors"
	"fmt"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"net"
	"time"
)

// Event 消息事件
type Event[T transport.Msg] struct {
	Flags transport.Flags // 标志位
	Msg   T               // 消息
}

// UnpackEvent 解包消息事件
func UnpackEvent[T transport.Msg](e Event[transport.Msg]) Event[T] {
	return Event[T]{
		Flags: e.Flags,
		Msg:   e.Msg.(T),
	}
}

// PackEvent 打包消息事件
func PackEvent[T transport.Msg](e Event[T]) Event[transport.Msg] {
	return Event[transport.Msg]{
		Flags: e.Flags,
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
	Conn    net.Conn       // 网络连接
	Encoder codec.IEncoder // 消息包编码器
	Decoder codec.IDecoder // 消息包解码器
	Timeout time.Duration  // io超时时间
}

// Send 发送消息事件
func (t *Transceiver) Send(e Event[transport.Msg]) error {
	if t.Conn == nil {
		return errors.New("conn is nil")
	}

	if t.Encoder == nil {
		return errors.New("encoder is nil")
	}

	if err := t.Encoder.Stuff(e.Flags, e.Msg); err != nil {
		return fmt.Errorf("stuff event msg failed, %w", err)
	}

	if t.Timeout > 0 {
		t.Conn.SetWriteDeadline(time.Now().Add(t.Timeout))
	} else {
		t.Conn.SetWriteDeadline(time.Time{})
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

	if err := t.Encoder.Stuff(transport.Flags_None(), msg); err != nil {
		return err
	}

	if t.Timeout > 0 {
		t.Conn.SetWriteDeadline(time.Now().Add(t.Timeout))
	} else {
		t.Conn.SetWriteDeadline(time.Time{})
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
			return Event[transport.Msg]{
				Flags: recvMP.Head.Flags,
				Msg:   recvMP.Msg,
			}, nil
		}

		if t.Timeout > 0 {
			t.Conn.SetReadDeadline(time.Now().Add(t.Timeout))
		} else {
			t.Conn.SetReadDeadline(time.Time{})
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
			b = fun(Event[transport.Msg]{
				Flags: mp.Head.Flags,
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

		if t.Timeout > 0 {
			t.Conn.SetReadDeadline(time.Now().Add(t.Timeout))
		} else {
			t.Conn.SetReadDeadline(time.Time{})
		}

		if _, err := t.Decoder.ReadFrom(t.Conn); err != nil {
			return fmt.Errorf("recv msg-packet failed, %w", err)
		}
	}
}
