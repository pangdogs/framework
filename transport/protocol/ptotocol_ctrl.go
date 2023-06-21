package protocol

import (
	"fmt"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"net"
	"time"
)

type (
	RecvRst       = func(Event[*transport.MsgRst]) error
	RecvSyncTime  = func(Event[*transport.MsgSyncTime]) error
	RecvHeartbeat = func(Event[*transport.MsgHeartbeat]) error
)

// CtrlProtocol 控制协议
type CtrlProtocol struct {
	Conn          net.Conn       // 网络连接
	Encoder       codec.IEncoder // 消息包编码器
	Timeout       time.Duration  // io超时时间
	RecvRst       RecvRst        // 接收Rst消息事件
	RecvSyncTime  RecvSyncTime   // 接收SyncTime消息事件
	RecvHeartbeat RecvHeartbeat  // 接收Heartbeat消息事件
}

// SendRst 发送Rst消息事件
func (c *CtrlProtocol) SendRst(err error) error {
	trans := Transceiver{
		Conn:    c.Conn,
		Encoder: c.Encoder,
		Timeout: c.Timeout,
	}
	return trans.SendRst(err)
}

// SendSyncTime 发送SyncTime消息事件
func (c *CtrlProtocol) SendSyncTime() error {
	trans := Transceiver{
		Conn:    c.Conn,
		Encoder: c.Encoder,
		Timeout: c.Timeout,
	}
	return trans.Send(PackEvent(Event[*transport.MsgSyncTime]{
		Msg: &transport.MsgSyncTime{UnixMilli: time.Now().UnixMilli()}},
	))
}

// SendHeartbeat 发送Heartbeat消息事件
func (c *CtrlProtocol) SendHeartbeat() error {
	trans := Transceiver{
		Conn:    c.Conn,
		Encoder: c.Encoder,
		Timeout: c.Timeout,
	}
	return trans.Send(PackEvent(Event[*transport.MsgHeartbeat]{
		Flags: transport.Flags(transport.Flag_Ping),
	}))
}

// Recv 消息事件处理句柄
func (c *CtrlProtocol) Recv(e Event[transport.Msg]) error {
	switch e.Msg.MsgId() {
	case transport.MsgId_Rst:
		if c.RecvRst != nil {
			return c.RecvRst(UnpackEvent[*transport.MsgRst](e))
		}
		return nil
	case transport.MsgId_SyncTime:
		if c.RecvSyncTime != nil {
			return c.RecvSyncTime(UnpackEvent[*transport.MsgSyncTime](e))
		}
		return nil
	case transport.MsgId_Heartbeat:
		heartbeat := UnpackEvent[*transport.MsgHeartbeat](e)

		if heartbeat.Flags.Is(transport.Flag_Ping) {
			trans := Transceiver{
				Conn:    c.Conn,
				Encoder: c.Encoder,
				Timeout: c.Timeout,
			}

			err := trans.Send(PackEvent(Event[*transport.MsgHeartbeat]{
				Flags: transport.Flags(transport.Flag_Pong),
			}))
			if err != nil {
				return err
			}
		}

		if c.RecvHeartbeat != nil {
			return c.RecvHeartbeat(heartbeat)
		}

		return nil
	default:
		return fmt.Errorf("%w: %d", ErrRecvUnexpectedMsg, e.Msg.MsgId())
	}
}
