package transport

import (
	"errors"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/net/gtp"
	"time"
)

type (
	RstHandler       = generic.DelegateFunc1[Event[gtp.MsgRst], error]       // Rst消息事件处理器
	SyncTimeHandler  = generic.DelegateFunc1[Event[gtp.MsgSyncTime], error]  // SyncTime消息事件处理器
	HeartbeatHandler = generic.DelegateFunc1[Event[gtp.MsgHeartbeat], error] // Heartbeat消息事件处理器
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
func (c *CtrlProtocol) RequestTime(corrId int64) error {
	if c.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	return c.retrySend(c.Transceiver.Send(
		Event[gtp.MsgSyncTime]{
			Flags: gtp.Flags(gtp.Flag_ReqTime),
			Msg: gtp.MsgSyncTime{
				CorrId:         corrId,
				LocalUnixMilli: time.Now().UnixMilli(),
			},
		}.Interface(),
	))
}

// SendPing 发送ping
func (c *CtrlProtocol) SendPing() error {
	if c.Transceiver == nil {
		return errors.New("setting Transceiver is nil")
	}
	return c.retrySend(c.Transceiver.Send(
		Event[gtp.MsgHeartbeat]{
			Flags: gtp.Flags(gtp.Flag_Ping),
		}.Interface(),
	))
}

func (c *CtrlProtocol) retrySend(err error) error {
	return Retry{
		Transceiver: c.Transceiver,
		Times:       c.RetryTimes,
	}.Send(err)
}

// HandleEvent 消息事件处理器
func (c *CtrlProtocol) HandleEvent(e IEvent) error {
	var errs []error

	interrupt := func(err, _ error) bool {
		if err != nil {
			errs = append(errs, err)
		}
		return false
	}

	switch e.Msg.MsgId() {
	case gtp.MsgId_Rst:
		c.RstHandler.Exec(interrupt, EventT[gtp.MsgRst](e))

	case gtp.MsgId_SyncTime:
		syncTime := EventT[gtp.MsgSyncTime](e)

		if syncTime.Flags.Is(gtp.Flag_ReqTime) {
			if c.Transceiver == nil {
				return errors.New("setting Transceiver is nil")
			}
			err := c.retrySend(c.Transceiver.Send(
				Event[gtp.MsgSyncTime]{
					Flags: gtp.Flags(gtp.Flag_RespTime),
					Msg: gtp.MsgSyncTime{
						CorrId:          syncTime.Msg.CorrId,
						LocalUnixMilli:  time.Now().UnixMilli(),
						RemoteUnixMilli: syncTime.Msg.LocalUnixMilli,
					},
				}.Interface(),
			))
			if err != nil {
				return err
			}
		}

		c.SyncTimeHandler.Exec(interrupt, syncTime)

	case gtp.MsgId_Heartbeat:
		heartbeat := EventT[gtp.MsgHeartbeat](e)

		if heartbeat.Flags.Is(gtp.Flag_Ping) {
			if c.Transceiver == nil {
				return errors.New("setting Transceiver is nil")
			}
			err := c.retrySend(c.Transceiver.Send(
				Event[gtp.MsgHeartbeat]{
					Flags: gtp.Flags(gtp.Flag_Pong),
				}.Interface(),
			))
			if err != nil {
				return err
			}
		}

		c.HeartbeatHandler.Exec(interrupt, heartbeat)

	default:
		return nil
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
