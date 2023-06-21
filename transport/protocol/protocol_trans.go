package protocol

import (
	"errors"
	"fmt"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"net"
	"time"
)

var (
	ErrRecvUnexpectedSeq = errors.New("recv unexpected payload seq") // 收到的数据序号错误
)

type (
	RecvPayload = func(Event[*transport.MsgPayload]) error
)

// TransProtocol 传输协议
type TransProtocol struct {
	Conn             net.Conn       // 网络连接
	Encoder          codec.IEncoder // 消息包编码器
	Timeout          time.Duration  // io超时时间
	SendSeq, RecvSeq uint32         // 请求响应序号
	RecvPayload      RecvPayload    // 接收Payload消息事件
}

// SendPayload 发送Payload消息事件
func (t *TransProtocol) SendPayload(e Event[*transport.MsgPayload]) error {
	trans := Transceiver{
		Conn:    t.Conn,
		Encoder: t.Encoder,
		Timeout: t.Timeout,
	}

	t.SendSeq++
	e.Msg.Seq = t.SendSeq

	return trans.Send(PackEvent(e))
}

// Recv 消息事件处理句柄
func (t *TransProtocol) Recv(e Event[transport.Msg]) error {
	switch e.Msg.MsgId() {
	case transport.MsgId_Payload:
		payload := UnpackEvent[*transport.MsgPayload](e)
		if payload.Msg.Seq != t.RecvSeq+1 {
			return ErrRecvUnexpectedSeq
		}

		t.RecvSeq++

		if t.RecvPayload != nil {
			return t.RecvPayload(payload)
		}

		return nil
	default:
		return fmt.Errorf("%w: %d", ErrRecvUnexpectedMsg, e.Msg.MsgId())
	}
}
