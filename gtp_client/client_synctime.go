package gtp_client

import (
	"context"
	"kit.golaxy.org/plugins/gtp/transport"
	"time"
)

// ResponseTime 响应同步时间
type ResponseTime struct {
	RequestTime time.Time // 请求时的本地时间
	LocalTime   time.Time // 响应时的本地时间
	RemoteTime  time.Time // 响应时的对端时间
}

// SyncTime 同步的时间
func (rt *ResponseTime) SyncTime() time.Time {
	return rt.RemoteTime.Add(rt.LocalTime.Sub(rt.RequestTime) / 2)
}

// RequestTime 请求对端同步时间
func (c *Client) RequestTime(ctx context.Context) <-chan transport.Ret[*ResponseTime] {
	resp := make(transport.AsyncRespChan[*ResponseTime], 1)

	reqId, err := c.asyncDispatcher.MakeRequest(ctx, resp)
	if err != nil {
		resp <- transport.Ret[*ResponseTime]{Error: err}
		return resp
	}

	err = c.ctrl.RequestTime(reqId)
	if err != nil {
		resp <- transport.Ret[*ResponseTime]{Error: err}
		return resp
	}

	return resp
}
