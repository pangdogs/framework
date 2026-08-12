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
	// ErrAutoReconnectRetriesExhausted 表示自动重连已用尽配置的尝试次数。
	ErrAutoReconnectRetriesExhausted = errors.New("cli: auto reconnect retries exhausted")
	// ErrInactiveTimeout 表示未启用自动重连的连接持续失活并超过等待时间。
	ErrInactiveTimeout = errors.New("cli: inactive timeout")
)

// NetAddr 保存客户端当前连接的本地与服务端地址快照。
type NetAddr struct {
	Local, Remote net.Addr // Local 是客户端地址，Remote 是服务端地址。
}

// Client 是支持 GTP 握手、心跳和连接迁移的并发安全客户端。
// 其上下文在父上下文取消或 Close 被调用时取消。
type Client struct {
	context.Context
	close            context.CancelCauseFunc
	closed           async.Completer
	options          ClientOptions
	sessionId        uid.Id
	endpoint         string
	netAddr          atomic.Pointer[NetAddr]
	transceiver      transport.Transceiver
	eventDispatcher  transport.EventDispatcher
	trans            transport.TransProtocol
	ctrl             transport.CtrlProtocol
	migrationMu      sync.Mutex
	migrationChan    chan struct{}
	migrations       atomic.Int64
	futureController *concurrent.FutureController
	io               _ClientIO
	logger           *zap.Logger
	sugarLogger      *zap.SugaredLogger
	stringerOnce     sync.Once
	stringerCache    string
}

// String 返回包含会话 ID 和用户 ID 的 JSON 形式标识。
func (c *Client) String() string {
	c.stringerOnce.Do(func() {
		c.stringerCache = fmt.Sprintf(`{"session_id":%q,"user_id":%q}`, c.SessionId(), c.UserId())
	})
	return c.stringerCache
}

// SessionId 返回服务端分配的会话 ID。
func (c *Client) SessionId() uid.Id {
	return c.sessionId
}

// UserId 返回握手使用的用户 ID。
func (c *Client) UserId() string {
	return c.options.AuthUserId
}

// Token 返回握手使用的鉴权令牌。
func (c *Client) Token() string {
	return c.options.AuthToken
}

// Extensions 返回握手使用的扩展数据；调用方不得修改。
func (c *Client) Extensions() []byte {
	return c.options.AuthExtensions
}

// Endpoint 返回建立连接时使用的服务端地址。
func (c *Client) Endpoint() string {
	return c.endpoint
}

// NetAddr 返回当前连接地址快照；连接迁移后会变化。
func (c *Client) NetAddr() NetAddr {
	return *c.netAddr.Load()
}

// Migrations 返回连接成功迁移的累计次数。
func (c *Client) Migrations() int64 {
	return c.migrations.Load()
}

// DataIO 返回原始负载 I/O 门面。
func (c *Client) DataIO() IDataIO {
	return (*_ClientDataIO)(&c.io)
}

// EventIO 返回 GTP 事件 I/O 门面。
func (c *Client) EventIO() IEventIO {
	return (*_ClientEventIO)(&c.io)
}

// FutureController 返回用于关联请求与响应的 Future 控制器。
func (c *Client) FutureController() *concurrent.FutureController {
	return c.futureController
}

// L 返回客户端结构化日志器。
func (c *Client) L() *zap.Logger {
	return c.logger
}

// S 返回客户端 SugaredLogger。
func (c *Client) S() *zap.SugaredLogger {
	return c.sugarLogger
}

// Close 请求以 err 为原因关闭客户端，并返回关闭完成信号。
func (c *Client) Close(err error) async.Signal {
	c.close(err)
	return c.closed.Signal()
}

// Closed 返回客户端关闭完成信号。
func (c *Client) Closed() async.Signal {
	return c.closed.Signal()
}

// handleHeartbeat 记录收到的 Ping 或 Pong 心跳事件。
func (c *Client) handleHeartbeat(event transport.Event[*gtp.MsgHeartbeat]) {
	if event.Flags.Is(gtp.Flag_Ping) {
		c.logger.Debug("client receive ping", zap.String("session_id", c.SessionId().String()))
	} else {
		c.logger.Debug("client receive pong", zap.String("session_id", c.SessionId().String()))
	}
}

// handleRst 将收到的 RST 事件转换为关闭原因并关闭客户端。
func (c *Client) handleRst(event transport.Event[*gtp.MsgRst]) {
	err := transport.ToRstError(event)
	c.logger.Debug("client receive rst",
		zap.String("session_id", c.SessionId().String()),
		zap.NamedError("rst_error", err))
	c.Close(err)
}
