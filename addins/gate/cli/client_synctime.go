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

// TimeSample 保存一次类 NTP 时钟探测的四个时间点，时间值以对端时区显示。
type TimeSample struct {
	OriginTime       time.Time // OriginTime 是请求方发送请求的 t1。
	ReceiveTime      time.Time // ReceiveTime 是响应方收到请求的 t2。
	TransmitTime     time.Time // TransmitTime 是响应方发送响应的 t3。
	DestinationTime  time.Time // DestinationTime 是请求方收到响应的 t4。
	RemoteZoneOffset int       // RemoteZoneOffset 是响应方相对 UTC 的偏移秒数。
}

// RTT 根据四个采样点估算扣除服务端处理时间后的网络往返时延。
func (ts TimeSample) RTT() time.Duration {
	return ts.DestinationTime.Sub(ts.OriginTime) - ts.TransmitTime.Sub(ts.ReceiveTime)
}

// Offset 估算对端时钟相对于本地时钟的偏移量。
func (ts TimeSample) Offset() time.Duration {
	return (ts.ReceiveTime.Sub(ts.OriginTime) + ts.TransmitTime.Sub(ts.DestinationTime)) / 2
}

// RemoteTime 估算请求方收到响应时刻的对端时间。
func (ts TimeSample) RemoteTime() time.Time {
	return ts.DestinationTime.Add(ts.Offset())
}

// RemoteNow 使用本次采样的固定偏移估算当前对端时间。
func (ts TimeSample) RemoteNow() time.Time {
	return time.Now().Add(ts.Offset())
}

// RemoteLocation 返回由采样时区偏移构造的固定时区。
func (ts TimeSample) RemoteLocation() *time.Location {
	return time.FixedZone("", ts.RemoteZoneOffset)
}

// ProbeTime 发起一次时钟探测，并返回承载 *TimeSample 的 Future。
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
