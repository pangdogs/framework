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
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework/net/gtp/codec"
	"git.golaxy.org/framework/utils/concurrent"
	"golang.org/x/net/websocket"
	"net"
	"net/url"
	"strings"
)

// _Connector 网络连接器
type _Connector struct {
	options        ClientOptions
	encoderCreator codec.EncoderCreator
	decoderCreator codec.DecoderCreator
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
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
		if err != nil {
			conn.Close()
		}
	}()

	client = ctor.newClient(ctx, conn, endpoint)

	err = ctor.handshake(ctx, conn, client)
	if err != nil {
		return nil, err
	}

	client.wg.Add(1)
	go client.mainLoop()

	return client, nil
}

// reconnect 重连服务端
func (ctor *_Connector) reconnect(client *Client) (err error) {
	if client == nil {
		return errors.New("client is nil")
	}

	select {
	case <-client.Done():
		return context.Canceled
	default:
		break
	}

	var conn net.Conn

	switch ctor.options.NetProtocol {
	case WebSocket:
		ep := client.GetEndpoint()

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
		conn, err = newDialer(&ctor.options).DialContext(client, "tcp", client.GetEndpoint())
		if err != nil {
			return err
		}

		if ctor.options.TLSConfig != nil {
			conn = tls.Client(conn, ctor.options.TLSConfig)
		}
	}

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
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
func (ctor *_Connector) newClient(ctx context.Context, conn net.Conn, endpoint string) *Client {
	client := &Client{
		options:  ctor.options,
		endpoint: endpoint,
		logger:   ctor.options.ZapLogger.Sugar(),
	}

	client.Context, client.terminate = context.WithCancelCause(ctx)
	client.terminated = async.MakeAsyncRet()
	client.transceiver.Conn = conn

	// 初始化消息事件分发器
	client.eventDispatcher.Transceiver = &client.transceiver
	client.eventDispatcher.RetryTimes = ctor.options.IORetryTimes
	client.eventDispatcher.EventHandler = generic.CastDelegate1(client.trans.HandleRecvEvent, client.ctrl.HandleRecvEvent, client.handleRecvEventChan, client.handleRecvEvent)

	// 初始化传输协议
	client.trans.Transceiver = &client.transceiver
	client.trans.RetryTimes = ctor.options.IORetryTimes
	client.trans.PayloadHandler = generic.CastDelegate1(client.handleRecvDataChan, client.handleRecvPayload)

	// 初始化控制协议
	client.ctrl.Transceiver = &client.transceiver
	client.ctrl.RetryTimes = ctor.options.IORetryTimes
	client.ctrl.HeartbeatHandler = generic.CastDelegate1(client.handleRecvHeartbeat)
	client.ctrl.SyncTimeHandler = generic.CastDelegate1(client.handleRecvSyncTime)
	client.ctrl.RstHandler = generic.CastDelegate1(client.handleRecvRst)

	// 初始化异步模型Future控制器
	client.futures = concurrent.NewFutures(client.Context, ctor.options.FutureTimeout)

	// 初始化监听器
	client.dataWatchers = concurrent.MakeLockedSlice[*_DataWatcher](0, 0)
	client.eventWatchers = concurrent.MakeLockedSlice[*_EventWatcher](0, 0)

	return client
}
