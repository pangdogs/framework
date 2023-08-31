package gtp_client

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/gtp/codec"
	"kit.golaxy.org/plugins/gtp/transport"
	"math/rand"
	"net"
)

// _Connector 网络连接器
type _Connector struct {
	Options ClientOptions
	encoder *codec.Encoder
	decoder *codec.Decoder
}

// Connect 连接服务端
func (ctor *_Connector) Connect(ctx context.Context, endpoint string) (client *Client, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	conn, err := newDialer(&ctor.Options).DialContext(ctx, "tcp", endpoint)
	if err != nil {
		return nil, err
	}

	if ctor.Options.TLSConfig != nil {
		conn = tls.Client(conn, ctor.Options.TLSConfig)
	}

	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("panicked: %w", panicErr)
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

// Reconnect 重连服务端
func (ctor *_Connector) Reconnect(client *Client) (err error) {
	if client == nil {
		return errors.New("client is nil")
	}

	conn, err := newDialer(&ctor.Options).DialContext(client, "tcp", client.GetEndpoint())
	if err != nil {
		return err
	}

	if ctor.Options.TLSConfig != nil {
		conn = tls.Client(conn, ctor.Options.TLSConfig)
	}

	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("panicked: %w", panicErr)
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
		options:  ctor.Options,
		endpoint: endpoint,
		logger:   ctor.Options.ZapLogger.Sugar(),
	}

	client.Context, client.cancel = context.WithCancel(ctx)
	client.transceiver.Conn = conn

	// 初始化消息事件分发器
	client.eventDispatcher.Transceiver = &client.transceiver
	client.eventDispatcher.RetryTimes = ctor.Options.IORetryTimes
	client.eventDispatcher.EventHandlers = []transport.EventHandler{client.trans.EventHandler, client.ctrl.EventHandler, client.eventHandler}

	// 初始化异步请求响应分发器
	client.asyncDispatcher.ReqId = rand.Int63()
	client.asyncDispatcher.Timeout = ctor.Options.IOTimeout

	// 初始化传输协议
	client.trans.Transceiver = &client.transceiver
	client.trans.RetryTimes = ctor.Options.IORetryTimes
	client.trans.PayloadHandler = client.payloadHandler

	// 初始化控制协议
	client.ctrl.Transceiver = &client.transceiver
	client.ctrl.RetryTimes = ctor.Options.IORetryTimes
	client.ctrl.HeartbeatHandler = client.heartbeatHandler
	client.ctrl.SyncTimeHandler = client.syncTimeHandler

	return client
}
