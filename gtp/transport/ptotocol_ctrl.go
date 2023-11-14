package transport

import (
	"errors"
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/plugins/gtp"
	"time"
)

type (
	RstHandler       = generic.DelegateFunc1[Event[*gtp.MsgRst], error]       // Rst消息事件处理器
	SyncTimeHandler  = generic.DelegateFunc1[Event[*gtp.MsgSyncTime], error]  // SyncTime消息事件处理器
	HeartbeatHandler = generic.DelegateFunc1[Event[*gtp.MsgHeartbeat], error] // Heartbeat消息事件处理器
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
			SeqId:          reqId,
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

// HandleEvent 消息事件处理器
func (c *CtrlProtocol) HandleEvent(e Event[gtp.Msg]) error {
	switch e.Msg.MsgId() {
	case gtp.MsgId_Rst:
		return c.RstHandler.Exec(func(err, _ error) bool {
			return err == nil || !errors.Is(err, ErrUnexpectedMsg)
		}, UnpackEvent[*gtp.MsgRst](e))

	case gtp.MsgId_SyncTime:
		syncTime := UnpackEvent[*gtp.MsgSyncTime](e)

		if syncTime.Flags.Is(gtp.Flag_ReqTime) {
			if c.Transceiver == nil {
				return errors.New("setting Transceiver is nil")
			}
			err := c.retrySend(c.Transceiver.Send(PackEvent(Event[*gtp.MsgSyncTime]{
				Flags: gtp.Flags(gtp.Flag_RespTime),
				Msg: &gtp.MsgSyncTime{
					SeqId:           syncTime.Msg.SeqId,
					LocalUnixMilli:  time.Now().UnixMilli(),
					RemoteUnixMilli: syncTime.Msg.LocalUnixMilli,
				},
			})))
			if err != nil {
				return err
			}
		}

		return c.SyncTimeHandler.Exec(func(err, _ error) bool {
			return err == nil || !errors.Is(err, ErrUnexpectedMsg)
		}, syncTime)

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

		return c.HeartbeatHandler.Exec(func(err, _ error) bool {
			return err == nil || !errors.Is(err, ErrUnexpectedMsg)
		}, heartbeat)

	default:
		return ErrUnexpectedMsg
	}
}
