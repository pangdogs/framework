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
	"context"
	"time"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"go.uber.org/zap"
)

// ResponseTime 响应同步时间
type ResponseTime struct {
	RequestTime time.Time // 请求时的本地时间
	LocalTime   time.Time // 响应时的本地时间
	RemoteTime  time.Time // 响应时的对端时间
}

// RTT 往返时间
func (rt ResponseTime) RTT() time.Duration {
	return rt.LocalTime.Sub(rt.RequestTime)
}

// SyncTime 同步的时间
func (rt ResponseTime) SyncTime() time.Time {
	return rt.RemoteTime.Add(rt.RTT() / 2)
}

// NowTime 当前时间
func (rt ResponseTime) NowTime() time.Time {
	return rt.SyncTime().Add(time.Now().Sub(rt.LocalTime))
}

// RequestTime 请求对端同步时间
func (c *Client) RequestTime(ctx context.Context) async.Future {
	handle := c.FutureController().New()
	if err := c.ctrl.RequestTime(handle.Id); err != nil {
		handle.Cancel(err)
	}
	return handle.Future()
}

// handleSyncTime 接收SyncTime消息事件
func (c *Client) handleSyncTime(event transport.Event[*gtp.MsgSyncTime]) {
	if event.Flags.Is(gtp.Flag_RespTime) {
		respTime := &ResponseTime{
			RequestTime: time.UnixMilli(event.Msg.RemoteTime).Local(),
			LocalTime:   time.Now(),
			RemoteTime:  time.UnixMilli(event.Msg.LocalTime).Local(),
		}
		err := c.futureController.Resolve(event.Msg.CorrId, async.NewResult(respTime, nil))
		if err != nil {
			c.logger.Error("failed to resolve future", zap.Int64("corr_id", event.Msg.CorrId), zap.Error(err))
		}
	}
}
