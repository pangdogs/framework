package protocol

import (
	"errors"
	"kit.golaxy.org/plugins/transport"
	"time"
)

type (
	HandleRst       = func(Event[*transport.MsgRst]) error
	HandleSyncTime  = func(Event[*transport.MsgSyncTime]) error
	HandleHeartbeat = func(Event[*transport.MsgHeartbeat]) error
)

// CtrlProtocol 控制协议
type CtrlProtocol struct {
	Transceiver     *Transceiver    // 消息事件收发器
	HandleRst       HandleRst       // Rst消息事件句柄
	HandleSyncTime  HandleSyncTime  // SyncTime消息事件句柄
	HandleHeartbeat HandleHeartbeat // Heartbeat消息事件句柄
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
	))
}

// SendPing 发送ping
func (c *CtrlProtocol) SendPing() error {
	if c.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	return c.Transceiver.Send(PackEvent(Event[*transport.MsgHeartbeat]{
		Flags: transport.Flags(transport.Flag_Sequenced | transport.Flag_Ping),
		Msg:   &transport.MsgHeartbeat{},
	}))
}

// HandleEvent 消息事件处理句柄
func (c *CtrlProtocol) HandleEvent(e Event[transport.Msg]) error {
	switch e.Msg.MsgId() {
	case transport.MsgId_Rst:
		if c.HandleRst != nil {
			return c.HandleRst(UnpackEvent[*transport.MsgRst](e))
		}
		return nil
	case transport.MsgId_SyncTime:
		if c.HandleSyncTime != nil {
			return c.HandleSyncTime(UnpackEvent[*transport.MsgSyncTime](e))
		}
		return nil
	case transport.MsgId_Heartbeat:
		heartbeat := UnpackEvent[*transport.MsgHeartbeat](e)

		if heartbeat.Flags.Is(transport.Flag_Ping) {
			if c.Transceiver == nil {
				return errors.New("setting Transceiver is nil")
			}
			err := c.Transceiver.Send(PackEvent(Event[*transport.MsgHeartbeat]{
				Flags: transport.Flags(transport.Flag_Sequenced | transport.Flag_Pong),
				Msg:   &transport.MsgHeartbeat{},
			}))
			if err != nil {
				return err
			}
		}

		if c.HandleHeartbeat != nil {
			return c.HandleHeartbeat(heartbeat)
		}

		return nil
	default:
		return ErrUnexpectedMsg
	}
}
