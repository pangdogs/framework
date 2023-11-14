package gtp_client

import (
	"context"
	"kit.golaxy.org/plugins/util/concurrent"
	"time"
)

// ResponseTime 响应同步时间
type ResponseTime struct {
	RequestTime time.Time // 请求时的本地时间
	LocalTime   time.Time // 响应时的本地时间
	RemoteTime  time.Time // 响应时的对端时间
}

// RTT 往返时间
func (rt *ResponseTime) RTT() time.Duration {
	return rt.LocalTime.Sub(rt.RequestTime)
}

// SyncTime 同步的时间
func (rt *ResponseTime) SyncTime() time.Time {
	return rt.RemoteTime.Add(rt.RTT() / 2)
}

// NowTime 当前时间
func (rt *ResponseTime) NowTime() time.Time {
	return rt.SyncTime().Add(time.Now().Sub(rt.LocalTime))
}

// RequestTime 请求对端同步时间
func (c *Client) RequestTime(ctx context.Context) <-chan concurrent.Ret[*ResponseTime] {
	resp := make(concurrent.RespChan[*ResponseTime], 1)
	future := c.futures.Make(ctx, resp)

	if err := c.ctrl.RequestTime(future.Id); err != nil {
		future.Cancel(err)
	}

	return resp
}
