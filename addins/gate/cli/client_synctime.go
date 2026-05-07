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

package cli

import (
	"time"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"go.uber.org/zap"
)

// TimeSample 时间采样，所有时间均按对端时区表示
type TimeSample struct {
	OriginTime       time.Time // NTP t1，请求方发送请求时间（对端时区）
	ReceiveTime      time.Time // NTP t2，响应方收到请求时间（对端时区）
	TransmitTime     time.Time // NTP t3，响应方发送响应时间（对端时区）
	DestinationTime  time.Time // NTP t4，请求方收到响应时间（对端时区）
	RemoteZoneOffset int       // 响应方时区偏移秒数
}

// RTT 往返时间
func (ts TimeSample) RTT() time.Duration {
	return ts.DestinationTime.Sub(ts.OriginTime) - ts.TransmitTime.Sub(ts.ReceiveTime)
}

// Offset 对端时间相对于本地时间的偏移量
func (ts TimeSample) Offset() time.Duration {
	return (ts.ReceiveTime.Sub(ts.OriginTime) + ts.TransmitTime.Sub(ts.DestinationTime)) / 2
}

// RemoteTime 估算请求方收到响应时刻的对端时间
func (ts TimeSample) RemoteTime() time.Time {
	return ts.DestinationTime.Add(ts.Offset())
}

// RemoteNow 估算当前对端时间
func (ts TimeSample) RemoteNow() time.Time {
	return time.Now().Add(ts.Offset())
}

// RemoteLocation 对端时区
func (ts TimeSample) RemoteLocation() *time.Location {
	return time.FixedZone("", ts.RemoteZoneOffset)
}

// ProbeTime 探测对端时间
func (c *Client) ProbeTime() async.Future {
	handle, err := c.FutureController().New()
	if err != nil {
		return async.Return(async.NewFutureChan(), async.NewResult(nil, err))
	}
	if err := c.ctrl.ProbeTime(handle.Id()); err != nil {
		handle.Cancel(err)
	}
	return handle.Future()
}

// handleSyncTime 接收SyncTime消息事件
func (c *Client) handleSyncTime(event transport.Event[*gtp.MsgSyncTime]) {
	if event.Flags.Is(gtp.Flag_RespTime) {
		remoteLocation := time.FixedZone("", int(event.Msg.ZoneOffset))
		timeSample := &TimeSample{
			OriginTime:       time.UnixMilli(event.Msg.OriginTime).In(remoteLocation),
			ReceiveTime:      time.UnixMilli(event.Msg.ReceiveTime).In(remoteLocation),
			TransmitTime:     time.UnixMilli(event.Msg.TransmitTime).In(remoteLocation),
			DestinationTime:  time.Now().In(remoteLocation),
			RemoteZoneOffset: int(event.Msg.ZoneOffset),
		}
		err := c.futureController.Resolve(event.Msg.CorrId, async.NewResult(timeSample, nil))
		if err != nil {
			c.logger.Error("failed to resolve future", zap.Int64("corr_id", event.Msg.CorrId), zap.Error(err))
		}
	}
}
