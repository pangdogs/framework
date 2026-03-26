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

// _Connector 网络连接器
type _Connector struct {
	options ClientOptions
	encoder *codec.Encoder
	decoder *codec.Decoder
}

// connect 连接服务端
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

		conn, err = websocket.DialConfig(conf)
		if err != nil {
			return nil, err
		}

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

// reconnect 重连服务端
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

		conn, err = websocket.DialConfig(conf)
		if err != nil {
			return err
		}

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

// newClient 创建客户端
func (ctor *_Connector) newClient(ctx context.Context, endpoint string) *Client {
	client := &Client{
		closed:        async.NewFutureVoid(),
		options:       ctor.options,
		endpoint:      endpoint,
		migrationChan: make(chan struct{}),
		logger:        ctor.options.Logger,
	}
	client.Context, client.close = context.WithCancelCause(ctx)

	// 初始化日志
	if client.Logger == nil {
		client.logger = zap.NewNop()
	}
	client.sugarLogger = client.logger.Sugar()

	// 初始化消息事件分发器
	client.eventDispatcher.AutoRecover = ctor.options.AutoRecover
	client.eventDispatcher.ReportError = ctor.options.ReportError
	client.eventDispatcher.Transceiver = &client.transceiver
	client.eventDispatcher.RetryTimes = ctor.options.IORetryTimes
	client.eventDispatcher.EventHandler = generic.CastDelegateVoid1(client.trans.HandleEvent, client.ctrl.HandleEvent, client.eventIO.handleEvent)

	// 初始化传输协议
	client.trans.AutoRecover = ctor.options.AutoRecover
	client.trans.ReportError = ctor.options.ReportError
	client.trans.Transceiver = &client.transceiver
	client.trans.RetryTimes = ctor.options.IORetryTimes
	client.trans.PayloadHandler = generic.CastDelegateVoid1(client.dataIO.handlePayload)

	// 初始化控制协议
	client.ctrl.AutoRecover = ctor.options.AutoRecover
	client.ctrl.ReportError = ctor.options.ReportError
	client.ctrl.Transceiver = &client.transceiver
	client.ctrl.RetryTimes = ctor.options.IORetryTimes
	client.ctrl.HeartbeatHandler = generic.CastDelegateVoid1(client.handleHeartbeat)
	client.ctrl.SyncTimeHandler = generic.CastDelegateVoid1(client.handleSyncTime)
	client.ctrl.RstHandler = generic.CastDelegateVoid1(client.handleRst)

	// 初始化异步模型Future控制器
	client.futureController = concurrent.NewFutureController(client.Context, ctor.options.FutureTimeout)

	// 初始化IO
	client.dataIO.init(client)
	client.eventIO.init(client)

	return client
}
