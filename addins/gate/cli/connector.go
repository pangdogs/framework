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
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework/net/gtp/codec"
	"git.golaxy.org/framework/utils/concurrent"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

// _Connector 按客户端选项建立 TCP 或 WebSocket 连接并完成 GTP 握手。
type _Connector struct {
	options ClientOptions
	encoder *codec.Encoder
	decoder *codec.Decoder
}

// connect 建立新客户端连接，握手成功后启动客户端主循环；失败时关闭底层连接。
func (ctor *_Connector) connect(ctx context.Context, endpoint string) (client *Client, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var conn net.Conn

	switch ctor.options.NetProtocol {
	case WebSocket:
		epURL, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}

		origin := ctor.options.WebSocketOrigin
		if origin == "" {
			origin, _ = url.JoinPath(endpoint, "cli", ctor.options.AuthUserId)
		}

		conf, err := websocket.NewConfig(endpoint, origin)
		if err != nil {
			return nil, err
		}

		if strings.EqualFold(epURL.Scheme, "https") || strings.EqualFold(epURL.Scheme, "wss") {
			if ctor.options.TLSConfig != nil {
				conf.TlsConfig = ctor.options.TLSConfig
			}
		}

		wsConn, err := conf.DialContext(ctx)
		if err != nil {
			return nil, err
		}

		wsConn.PayloadType = websocket.BinaryFrame
		conn = wsConn

	default:
		conn, err = newDialer(&ctor.options).DialContext(ctx, "tcp", endpoint)
		if err != nil {
			return nil, err
		}

		if ctor.options.TLSConfig != nil {
			conn = tls.Client(conn, ctor.options.TLSConfig)
		}
	}

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("cli: %w: %w", core.ErrPanicked, panicErr)
		}
		if err != nil {
			conn.Close()
		}
	}()

	client = ctor.newClient(ctx, endpoint)

	err = ctor.handshake(ctx, conn, client)
	if err != nil {
		return nil, err
	}

	go client.mainLoop()

	return client, nil
}

// reconnect 为现有客户端建立新连接并执行会话续接；失败时关闭新连接。
func (ctor *_Connector) reconnect(client *Client) (err error) {
	if client == nil {
		return errors.New("cli: client is nil")
	}

	select {
	case <-client.Done():
		return client.Err()
	default:
		break
	}

	var conn net.Conn

	switch ctor.options.NetProtocol {
	case WebSocket:
		ep := client.Endpoint()

		epURL, err := url.Parse(ep)
		if err != nil {
			return err
		}

		origin := ctor.options.WebSocketOrigin
		if origin == "" {
			origin, _ = url.JoinPath(ep, "cli", ctor.options.AuthUserId)
		}

		conf, err := websocket.NewConfig(ep, origin)
		if err != nil {
			return err
		}

		if strings.EqualFold(epURL.Scheme, "https") || strings.EqualFold(epURL.Scheme, "wss") {
			if ctor.options.TLSConfig != nil {
				conf.TlsConfig = ctor.options.TLSConfig
			}
		}

		wsConn, err := conf.DialContext(client)
		if err != nil {
			return err
		}

		wsConn.PayloadType = websocket.BinaryFrame
		conn = wsConn

	default:
		conn, err = newDialer(&ctor.options).DialContext(client, "tcp", client.Endpoint())
		if err != nil {
			return err
		}

		if ctor.options.TLSConfig != nil {
			conn = tls.Client(conn, ctor.options.TLSConfig)
		}
	}

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("cli: %w: %w", core.ErrPanicked, panicErr)
		}
		if err != nil {
			conn.Close()
		}
	}()

	err = ctor.handshake(client, conn, client)
	if err != nil {
		return err
	}

	return nil
}

// newClient 创建尚未绑定连接的客户端，并装配传输、控制、Future 和 I/O 处理器。
func (ctor *_Connector) newClient(ctx context.Context, endpoint string) *Client {
	client := &Client{
		closed:        async.NewFutureVoid(),
		options:       ctor.options,
		endpoint:      endpoint,
		migrationChan: make(chan struct{}),
		logger:        ctor.options.Logger,
	}
	client.Context, client.close = context.WithCancelCause(ctx)

	// 未配置日志器时使用静默日志器，避免在热路径反复判空。
	if client.logger == nil {
		client.logger = zap.NewNop()
	}
	client.sugarLogger = client.logger.Sugar()

	// 分发器按传输、控制、业务事件的顺序尝试处理同一事件。
	client.eventDispatcher.AutoRecover = ctor.options.AutoRecover
	client.eventDispatcher.ReportError = ctor.options.ReportError
	client.eventDispatcher.Transceiver = &client.transceiver
	client.eventDispatcher.RetryTimes = ctor.options.IORetryTimes
	client.eventDispatcher.EventHandler = generic.CastDelegateVoid1(client.trans.HandleEvent, client.ctrl.HandleEvent, client.io.handleEvent)

	// 传输协议负责业务负载与重试。
	client.trans.AutoRecover = ctor.options.AutoRecover
	client.trans.ReportError = ctor.options.ReportError
	client.trans.Transceiver = &client.transceiver
	client.trans.RetryTimes = ctor.options.IORetryTimes
	client.trans.PayloadHandler = generic.CastDelegateVoid1(client.io.handlePayload)

	// 控制协议负责心跳、时钟同步和链路重置事件。
	client.ctrl.AutoRecover = ctor.options.AutoRecover
	client.ctrl.ReportError = ctor.options.ReportError
	client.ctrl.Transceiver = &client.transceiver
	client.ctrl.RetryTimes = ctor.options.IORetryTimes
	client.ctrl.HeartbeatHandler = generic.CastDelegateVoid1(client.handleHeartbeat)
	client.ctrl.SyncTimeHandler = generic.CastDelegateVoid1(client.handleSyncTime)
	client.ctrl.RstHandler = generic.CastDelegateVoid1(client.handleRst)

	// Future 控制器随客户端上下文取消，统一终止未完成请求。
	client.futureController = concurrent.NewFutureController(client.Context, ctor.options.FutureTimeout)

	// I/O 门面使用独立队列异步发送数据和事件。
	client.io.init(client)

	return client
}
