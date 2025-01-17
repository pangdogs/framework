/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package transport

import (
	"errors"
	"fmt"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/net/gtp"
	"time"
)

type (
	RstHandler       = generic.Delegate1[Event[gtp.MsgRst], error]       // Rst消息事件处理器
	SyncTimeHandler  = generic.Delegate1[Event[gtp.MsgSyncTime], error]  // SyncTime消息事件处理器
	HeartbeatHandler = generic.Delegate1[Event[gtp.MsgHeartbeat], error] // Heartbeat消息事件处理器
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
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}

	// rst消息不重试
	retErr := c.Transceiver.SendRst(err)
	if retErr != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, retErr)
	}

	return nil
}

// RequestTime 请求同步时间
func (c *CtrlProtocol) RequestTime(corrId int64) error {
	if c.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}

	err := c.retrySend(c.Transceiver.Send(
		Event[gtp.MsgSyncTime]{
			Flags: gtp.Flags(gtp.Flag_ReqTime),
			Msg: gtp.MsgSyncTime{
				CorrId:         corrId,
				LocalUnixMilli: time.Now().UnixMilli(),
			},
		}.Interface(),
	))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// SendPing 发送ping
func (c *CtrlProtocol) SendPing() error {
	if c.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}

	err := c.retrySend(c.Transceiver.Send(
		Event[gtp.MsgHeartbeat]{
			Flags: gtp.Flags(gtp.Flag_Ping),
		}.Interface(),
	))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
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
		c.RstHandler.UnsafeCall(interrupt, EventT[gtp.MsgRst](e))

	case gtp.MsgId_SyncTime:
		syncTime := EventT[gtp.MsgSyncTime](e)

		if syncTime.Flags.Is(gtp.Flag_ReqTime) {
			if c.Transceiver == nil {
				return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
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

		c.SyncTimeHandler.UnsafeCall(interrupt, syncTime)

	case gtp.MsgId_Heartbeat:
		heartbeat := EventT[gtp.MsgHeartbeat](e)

		if heartbeat.Flags.Is(gtp.Flag_Ping) {
			if c.Transceiver == nil {
				return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
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

		c.HeartbeatHandler.UnsafeCall(interrupt, heartbeat)

	default:
		return nil
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
