package protocol

import (
	"fmt"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"net"
)

type (
	RecvPayload = func(Event[*transport.MsgPayload]) error
)

// Trans 传输协议
type Trans struct {
	Conn        net.Conn       // 网络连接
	Encoder     codec.IEncoder // 消息包编码器
	RetryTimes  int            // io超时重试次数
	RecvPayload RecvPayload    // 接收Payload消息事件
}

// SendPayload 发送Payload消息事件
func (t *Trans) SendPayload(e Event[*transport.MsgPayload]) error {
	trans := Transceiver{
		Conn:       t.Conn,
		Encoder:    t.Encoder,
		RetryTimes: t.RetryTimes,
	}
	return trans.Send(PackEvent(e))
}

// Recv 消息事件处理句柄
func (t *Trans) Recv(e Event[transport.Msg]) error {
	switch e.Msg.MsgId() {
	case transport.MsgId_Payload:
		if t.RecvPayload != nil {
			return t.RecvPayload(UnpackEvent[*transport.MsgPayload](e))
		}
		return nil
	default:
		return fmt.Errorf("recv unexpected msg %d", e.Msg.MsgId())
	}
}
