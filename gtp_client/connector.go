package gtp_client

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/gtp/codec"
	"math/rand"
	"net"
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

	conn, err := newDialer(&ctor.options).DialContext(ctx, "tcp", endpoint)
	if err != nil {
		return nil, err
	}

	if ctor.options.TLSConfig != nil {
		conn = tls.Client(conn, ctor.options.TLSConfig)
	}

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", golaxy.ErrPanicked, panicErr)
		}
		if err != nil {
			conn.Close()
		}
	}()

	client = ctor.newClient(ctx, conn, endpoint)

	err = ctor.handshake(conn, client)
	if err != nil {
		return nil, err
	}

	go client.run()

	return client, nil
}

// reconnect 重连服务端
func (ctor *_Connector) reconnect(client *Client) (err error) {
	if client == nil {
		return errors.New("client is nil")
	}

	conn, err := newDialer(&ctor.options).DialContext(client, "tcp", client.GetEndpoint())
	if err != nil {
		return err
	}

	if ctor.options.TLSConfig != nil {
		conn = tls.Client(conn, ctor.options.TLSConfig)
	}

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", golaxy.ErrPanicked, panicErr)
		}
		if err != nil {
			conn.Close()
		}
	}()

	err = ctor.handshake(conn, client)
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

	client.Context, client.cancel = context.WithCancel(ctx)
	client.closedChan = make(chan struct{}, 1)
	client.transceiver.Conn = conn

	// 初始化消息事件分发器
	client.eventDispatcher.Transceiver = &client.transceiver
	client.eventDispatcher.RetryTimes = ctor.options.IORetryTimes
	client.eventDispatcher.EventHandler = generic.CastDelegateFunc1(client.trans.HandleEvent, client.ctrl.HandleEvent, client.handleEvent)

	// 初始化传输协议
	client.trans.Transceiver = &client.transceiver
	client.trans.RetryTimes = ctor.options.IORetryTimes
	client.trans.PayloadHandler = generic.CastDelegateFunc1(client.handlePayload)

	// 初始化控制协议
	client.ctrl.Transceiver = &client.transceiver
	client.ctrl.RetryTimes = ctor.options.IORetryTimes
	client.ctrl.HeartbeatHandler = generic.CastDelegateFunc1(client.handleHeartbeat)
	client.ctrl.SyncTimeHandler = generic.CastDelegateFunc1(client.handleSyncTime)

	// 初始化异步模型Future控制器
	client.futures.Ctx = client.Context
	client.futures.Id = rand.Int63()
	client.futures.Timeout = ctor.options.FutureTimeout

	return client
}
