package transport

import (
	"errors"
	"kit.golaxy.org/plugins/gtp"
	"time"
)

type (
	RstHandler       = func(Event[*gtp.MsgRst]) error       // Rst消息事件处理器
	SyncTimeHandler  = func(Event[*gtp.MsgSyncTime]) error  // SyncTime消息事件处理器
	HeartbeatHandler = func(Event[*gtp.MsgHeartbeat]) error // Heartbeat消息事件处理器
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

// RequestTime 请求同步时间
func (c *CtrlProtocol) RequestTime(reqId int64) error {
	if c.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	return c.retrySend(c.Transceiver.Send(PackEvent(Event[*gtp.MsgSyncTime]{
		Flags: gtp.Flags(gtp.Flag_ReqTime),
		Msg: &gtp.MsgSyncTime{
			ReqId:          reqId,
			LocalUnixMilli: time.Now().UnixMilli(),
		}},
	)))
}

// SendPing 发送ping
func (c *CtrlProtocol) SendPing() error {
	if c.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	return c.retrySend(c.Transceiver.Send(PackEvent(Event[*gtp.MsgHeartbeat]{
		Flags: gtp.Flags(gtp.Flag_Ping),
		Msg:   &gtp.MsgHeartbeat{},
	})))
}

func (c *CtrlProtocol) retrySend(err error) error {
	return Retry{
		Transceiver: c.Transceiver,
		Times:       c.RetryTimes,
	}.Send(err)
}

// EventHandler 消息事件处理器
func (c *CtrlProtocol) EventHandler(e Event[gtp.Msg]) error {
	switch e.Msg.MsgId() {
	case gtp.MsgId_Rst:
		if c.RstHandler != nil {
			return c.RstHandler(UnpackEvent[*gtp.MsgRst](e))
		}
		return nil
	case gtp.MsgId_SyncTime:
		syncTime := UnpackEvent[*gtp.MsgSyncTime](e)

		if syncTime.Flags.Is(gtp.Flag_ReqTime) {
			if c.Transceiver == nil {
				return errors.New("setting Transceiver is nil")
			}
			err := c.retrySend(c.Transceiver.Send(PackEvent(Event[*gtp.MsgSyncTime]{
				Flags: gtp.Flags(gtp.Flag_RespTime),
				Msg: &gtp.MsgSyncTime{
					ReqId:           syncTime.Msg.ReqId,
					LocalUnixMilli:  time.Now().UnixMilli(),
					RemoteUnixMilli: syncTime.Msg.LocalUnixMilli,
				},
			})))
			if err != nil {
				return err
			}
		}

		if c.SyncTimeHandler != nil {
			return c.SyncTimeHandler(syncTime)
		}

		return nil
	case gtp.MsgId_Heartbeat:
		heartbeat := UnpackEvent[*gtp.MsgHeartbeat](e)

		if heartbeat.Flags.Is(gtp.Flag_Ping) {
			if c.Transceiver == nil {
				return errors.New("setting Transceiver is nil")
			}
			err := c.retrySend(c.Transceiver.Send(PackEvent(Event[*gtp.MsgHeartbeat]{
				Flags: gtp.Flags(gtp.Flag_Pong),
				Msg:   &gtp.MsgHeartbeat{},
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
