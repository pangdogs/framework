package protocol

import (
	"errors"
	"kit.golaxy.org/plugins/transport"
	"time"
)

type (
	RstHandler       = func(Event[*transport.MsgRst]) error       // Rst消息事件处理器
	SyncTimeHandler  = func(Event[*transport.MsgSyncTime]) error  // SyncTime消息事件处理器
	HeartbeatHandler = func(Event[*transport.MsgHeartbeat]) error // Heartbeat消息事件处理器
)

// CtrlProtocol 控制协议
type CtrlProtocol struct {
	Transceiver      *Transceiver     // 消息事件收发器
	RetryTimes       int              // 网络io超时时的重试次数
	RstHandler       RstHandler       // Rst消息事件处理器
	SyncTimeHandler  SyncTimeHandler  // SyncTime消息事件处理器
	HeartbeatHandler HeartbeatHandler // Heartbeat消息事件处理器
}

// SendRst 发送Rst消息事件
func (c *CtrlProtocol) SendRst(err error) error {
	if c.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	// rst消息不重试
	return c.Transceiver.SendRst(err)
}

// SendSyncTime 发送SyncTime消息事件
func (c *CtrlProtocol) SendSyncTime() error {
	if c.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	return c.retrySend(c.Transceiver.Send(PackEvent(Event[*transport.MsgSyncTime]{
		Msg: &transport.MsgSyncTime{UnixMilli: time.Now().UnixMilli()}},
	)))
}

// SendPing 发送ping
func (c *CtrlProtocol) SendPing() error {
	if c.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	return c.retrySend(c.Transceiver.Send(PackEvent(Event[*transport.MsgHeartbeat]{
		Flags: transport.Flags(transport.Flag_Sequenced | transport.Flag_Ping),
		Msg:   &transport.MsgHeartbeat{},
	})))
}

func (c *CtrlProtocol) retrySend(err error) error {
	return Retry{
		Transceiver: c.Transceiver,
		Times:       c.RetryTimes,
	}.Send(err)
}

// EventHandler 消息事件处理器
func (c *CtrlProtocol) EventHandler(e Event[transport.Msg]) error {
	switch e.Msg.MsgId() {
	case transport.MsgId_Rst:
		if c.RstHandler != nil {
			return c.RstHandler(UnpackEvent[*transport.MsgRst](e))
		}
		return nil
	case transport.MsgId_SyncTime:
		if c.SyncTimeHandler != nil {
			return c.SyncTimeHandler(UnpackEvent[*transport.MsgSyncTime](e))
		}
		return nil
	case transport.MsgId_Heartbeat:
		heartbeat := UnpackEvent[*transport.MsgHeartbeat](e)

		if heartbeat.Flags.Is(transport.Flag_Ping) {
			if c.Transceiver == nil {
				return errors.New("setting Transceiver is nil")
			}
			err := c.retrySend(c.Transceiver.Send(PackEvent(Event[*transport.MsgHeartbeat]{
				Flags: transport.Flags(transport.Flag_Sequenced | transport.Flag_Pong),
				Msg:   &transport.MsgHeartbeat{},
			})))
			if err != nil {
				return err
			}
		}

		if c.HeartbeatHandler != nil {
			return c.HeartbeatHandler(heartbeat)
		}

		return nil
	default:
		return ErrUnexpectedMsg
	}
}
