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
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/concurrent"
	"go.uber.org/zap"
)

var (
	ErrAutoReconnectRetriesExhausted = errors.New("cli: auto reconnect retries exhausted")
	ErrInactiveTimeout               = errors.New("cli: inactive timeout")
)

// NetAddr 网络地址
type NetAddr struct {
	Local, Remote net.Addr
}

// Client 客户端
type Client struct {
	context.Context
	close            context.CancelCauseFunc
	closed           async.FutureVoid
	options          ClientOptions
	sessionId        uid.Id
	endpoint         string
	netAddr          atomic.Pointer[NetAddr]
	transceiver      transport.Transceiver
	eventDispatcher  transport.EventDispatcher
	trans            transport.TransProtocol
	ctrl             transport.CtrlProtocol
	migrationMutex   sync.Mutex
	migrationChan    chan struct{}
	migrations       atomic.Int64
	futureController *concurrent.FutureController
	io               _ClientIO
	logger           *zap.Logger
	sugarLogger      *zap.SugaredLogger
	stringerOnce     sync.Once
	stringerCache    string
}

// String implements fmt.Stringer
func (c *Client) String() string {
	c.stringerOnce.Do(func() {
		c.stringerCache = fmt.Sprintf(`{"session_id":%q,"user_id":%q}`, c.SessionId(), c.UserId())
	})
	return c.stringerCache
}

// SessionId 获取会话Id
func (c *Client) SessionId() uid.Id {
	return c.sessionId
}

// UserId 获取鉴权用户Id
func (c *Client) UserId() string {
	return c.options.AuthUserId
}

// Token 获取鉴权token
func (c *Client) Token() string {
	return c.options.AuthToken
}

// Extensions 获取鉴权扩展数据
func (c *Client) Extensions() []byte {
	return c.options.AuthExtensions
}

// Endpoint 获取服务器地址
func (c *Client) Endpoint() string {
	return c.endpoint
}

// NetAddr 获取网络地址
func (c *Client) NetAddr() NetAddr {
	return *c.netAddr.Load()
}

// Migrations 获取连接迁移次数
func (c *Client) Migrations() int64 {
	return c.migrations.Load()
}

// DataIO 获取数据IO
func (c *Client) DataIO() IDataIO {
	return (*_ClientDataIO)(&c.io)
}

// EventIO 获取事件IO
func (c *Client) EventIO() IEventIO {
	return (*_ClientEventIO)(&c.io)
}

// FutureController 获取异步模型Future控制器
func (c *Client) FutureController() *concurrent.FutureController {
	return c.futureController
}

// Logger 获取logger
func (c *Client) Logger() *zap.Logger {
	return c.logger
}

// SugarLogger 获取SugarLogger
func (c *Client) SugarLogger() *zap.SugaredLogger {
	return c.sugarLogger
}

// Close 关闭
func (c *Client) Close(err error) async.Future {
	c.close(err)
	return c.closed.Out()
}

// Closed 已关闭
func (c *Client) Closed() async.Future {
	return c.closed.Out()
}

// handleHeartbeat 接收Heartbeat消息事件
func (c *Client) handleHeartbeat(event transport.Event[*gtp.MsgHeartbeat]) {
	if event.Flags.Is(gtp.Flag_Ping) {
		c.logger.Debug("client receive ping", zap.String("session_id", c.SessionId().String()))
	} else {
		c.logger.Debug("client receive pong", zap.String("session_id", c.SessionId().String()))
	}
}

// handleRst 接收Rst消息事件
func (c *Client) handleRst(event transport.Event[*gtp.MsgRst]) {
	err := transport.CastRstErr(event)
	c.logger.Debug("client receive rst",
		zap.String("session_id", c.SessionId().String()),
		zap.NamedError("rst_error", err))
	c.Close(err)
}
