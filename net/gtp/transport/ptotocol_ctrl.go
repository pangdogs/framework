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
	"fmt"
	"time"

	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/net/gtp"
)

type (
	// RstHandler 处理链路重置事件。
	RstHandler = generic.DelegateVoid1[Event[*gtp.MsgRst]]
	// SyncTimeHandler 处理时钟同步请求或响应事件。
	SyncTimeHandler = generic.DelegateVoid1[Event[*gtp.MsgSyncTime]]
	// HeartbeatHandler 处理心跳探测或响应事件。
	HeartbeatHandler = generic.DelegateVoid1[Event[*gtp.MsgHeartbeat]]
)

// CtrlProtocol 发送并处理链路重置、时钟同步和心跳控制消息。
type CtrlProtocol struct {
	AutoRecover      bool             // 是否恢复控制消息处理器的 panic。
	ReportError      chan error       // 恢复 panic 后接收错误；nil 时不报告。
	Transceiver      *Transceiver     // 事件收发器。
	RetryTimes       int              // 网络 I/O 超时后的重试次数。
	RstHandler       RstHandler       // 链路重置处理器。
	SyncTimeHandler  SyncTimeHandler  // 时钟同步处理器。
	HeartbeatHandler HeartbeatHandler // 心跳处理器。
}

// SendRst 发送不重试的链路重置事件。
func (c *CtrlProtocol) SendRst(err error) error {
	if c.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}

	// 链路已异常时重试 RST 没有可靠意义。
	retErr := c.Transceiver.SendRst(err)
	if retErr != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, retErr)
	}

	return nil
}

// SendPing 发送心跳探测，并在网络 I/O 超时时重试。
func (c *CtrlProtocol) SendPing() error {
	if c.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}

	err := c.retrySend(c.Transceiver.Send(
		Event[*gtp.MsgHeartbeat]{
			Flags: gtp.Flags(gtp.Flag_Ping),
			Msg:   &gtp.MsgHeartbeat{},
		}.Interface(),
	))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProtocol, err)
	}

	return nil
}

// ProbeTime 发送带关联 ID 和本地发送时间的时钟同步请求。
func (c *CtrlProtocol) ProbeTime(corrId int64) error {
	if c.Transceiver == nil {
		return fmt.Errorf("%w: Transceiver is nil", ErrProtocol)
	}

	err := c.retrySend(c.Transceiver.Send(
		Event[*gtp.MsgSyncTime]{
			Flags: gtp.Flags(gtp.Flag_ReqTime),
			Msg: &gtp.MsgSyncTime{
				CorrId:     corrId,
				OriginTime: time.Now().UnixMilli(),
			},
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

// HandleEvent 自动响应时间请求和心跳探测，再同步调用对应处理器。
func (c *CtrlProtocol) HandleEvent(e IEvent) {
	switch e.Msg.MsgId() {
	case gtp.MsgId_Rst:
		c.RstHandler.Call(c.AutoRecover, c.ReportError, nil, AssertEvent[*gtp.MsgRst](e))

	case gtp.MsgId_SyncTime:
		syncTime := AssertEvent[*gtp.MsgSyncTime](e)

		if syncTime.Flags.Is(gtp.Flag_ReqTime) {
			recvTime := time.Now()
			_, zoneOffset := recvTime.Zone()
			if c.Transceiver == nil {
				exception.Panicf("%w: Transceiver is nil", ErrProtocol)
			}
			err := c.retrySend(c.Transceiver.Send(
				Event[*gtp.MsgSyncTime]{
					Flags: gtp.Flags(gtp.Flag_RespTime),
					Msg: &gtp.MsgSyncTime{
						CorrId:       syncTime.Msg.CorrId,
						OriginTime:   syncTime.Msg.OriginTime,
						ReceiveTime:  recvTime.UnixMilli(),
						TransmitTime: time.Now().UnixMilli(),
						ZoneOffset:   int32(zoneOffset),
					},
				}.Interface(),
			))
			if err != nil {
				exception.Panicf("%w: %w", ErrProtocol, err)
			}
		}

		c.SyncTimeHandler.Call(c.AutoRecover, c.ReportError, nil, syncTime)

	case gtp.MsgId_Heartbeat:
		heartbeat := AssertEvent[*gtp.MsgHeartbeat](e)

		if heartbeat.Flags.Is(gtp.Flag_Ping) {
			if c.Transceiver == nil {
				exception.Panicf("%w: Transceiver is nil", ErrProtocol)
			}
			err := c.retrySend(c.Transceiver.Send(
				Event[*gtp.MsgHeartbeat]{
					Flags: gtp.Flags(gtp.Flag_Pong),
					Msg:   &gtp.MsgHeartbeat{},
				}.Interface(),
			))
			if err != nil {
				exception.Panicf("%w: %w", ErrProtocol, err)
			}
		}

		c.HeartbeatHandler.Call(c.AutoRecover, c.ReportError, nil, heartbeat)
	}
}
