package protocol

import (
	"errors"
	"fmt"
	"kit.golaxy.org/plugins/transport"
	"time"
)

type (
	RecvRst       = func(Event[*transport.MsgRst]) error
	RecvSyncTime  = func(Event[*transport.MsgSyncTime]) error
	RecvHeartbeat = func(Event[*transport.MsgHeartbeat]) error
)

// CtrlProtocol 控制协议
type CtrlProtocol struct {
	Transceiver   *Transceiver  // 消息事件收发器
	RecvRst       RecvRst       // 接收Rst消息事件
	RecvSyncTime  RecvSyncTime  // 接收SyncTime消息事件
	RecvHeartbeat RecvHeartbeat // 接收Heartbeat消息事件
}

// SendRst 发送Rst消息事件
func (c *CtrlProtocol) SendRst(err error) error {
	if c.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	return c.Transceiver.SendRst(err)
}

// SendSyncTime 发送SyncTime消息事件
func (c *CtrlProtocol) SendSyncTime() error {
	if c.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	return c.Transceiver.Send(PackEvent(Event[*transport.MsgSyncTime]{
		Msg: &transport.MsgSyncTime{UnixMilli: time.Now().UnixMilli()}},
	), false)
}

// SendHeartbeat 发送Heartbeat消息事件
func (c *CtrlProtocol) SendHeartbeat() error {
	if c.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	return c.Transceiver.Send(PackEvent(Event[*transport.MsgHeartbeat]{
		Flags: transport.Flags(transport.Flag_Ping),
	}), false)
}

// Bind 绑定事件分发器
func (c *CtrlProtocol) Bind(dispatcher *Dispatcher) error {
	if dispatcher == nil {
		return errors.New("dispatcher is nil")
	}

	if dispatcher.Handlers == nil {
		dispatcher.Handlers = map[transport.MsgId]Handler{}
	}

	dispatcher.Handlers[transport.MsgId_Rst] = c
	dispatcher.Handlers[transport.MsgId_SyncTime] = c
	dispatcher.Handlers[transport.MsgId_Heartbeat] = c

	return nil
}

// Unbind 解绑定事件分发器
func (c *CtrlProtocol) Unbind(dispatcher *Dispatcher) error {
	if dispatcher == nil {
		return errors.New("dispatcher is nil")
	}

	if dispatcher.Handlers == nil {
		return nil
	}

	if dispatcher.Handlers[transport.MsgId_Rst] == c {
		delete(dispatcher.Handlers, transport.MsgId_Rst)
	}
	if dispatcher.Handlers[transport.MsgId_SyncTime] == c {
		delete(dispatcher.Handlers, transport.MsgId_SyncTime)
	}
	if dispatcher.Handlers[transport.MsgId_Heartbeat] == c {
		delete(dispatcher.Handlers, transport.MsgId_Heartbeat)
	}

	return nil
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
			if c.Transceiver == nil {
				return errors.New("setting Transceiver is nil")
			}
			err := c.Transceiver.Send(PackEvent(Event[*transport.MsgHeartbeat]{
				Flags: transport.Flags(transport.Flag_Pong),
			}), false)
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
