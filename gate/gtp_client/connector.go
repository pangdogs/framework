package gtp_client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/transport/protocol"
)

// _Connector 网络连接器
type _Connector struct {
	Options ClientOptions
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

	ctx, cancel := context.WithCancel(ctx)

	client = &Client{
		Context:  ctx,
		cancel:   cancel,
		options:  ctor.Options,
		endpoint: endpoint,
		logger:   ctor.Options.ZapLogger.Sugar(),
	}

	client.transceiver.Conn = conn

	// 初始化消息事件分发器
	client.dispatcher.Transceiver = &client.transceiver
	client.dispatcher.EventHandlers = []protocol.EventHandler{client.trans.EventHandler, client.ctrl.EventHandler, client.eventHandler}

	// 初始化传输协议
	client.trans.Transceiver = &client.transceiver
	client.trans.PayloadHandler = client.payloadHandler

	// 初始化控制协议
	client.ctrl.Transceiver = &client.transceiver

	err = ctor.handshake(conn, client)
	if err != nil {
		return nil, err
	}

	go client.run()

	return client, nil
}

// Reconnect 重连服务端
func (ctor *_Connector) Reconnect(client *Client) error {
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

	err = ctor.handshake(conn, client)
	if err != nil {
		return err
	}

	return nil
}
