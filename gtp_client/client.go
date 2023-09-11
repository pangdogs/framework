package gtp_client

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"kit.golaxy.org/plugins/gtp"
	"kit.golaxy.org/plugins/gtp/transport"
	"net"
	"sync"
)

// Client 客户端
type Client struct {
	context.Context
	cancel          context.CancelFunc
	mutex           sync.Mutex
	options         ClientOptions
	sessionId       string
	endpoint        string
	transceiver     transport.Transceiver
	eventDispatcher transport.EventDispatcher
	trans           transport.TransProtocol
	ctrl            transport.CtrlProtocol
	reconnectChan   chan struct{}
	renewChan       chan struct{}
	promise         transport.Promise
	logger          *zap.SugaredLogger
}

// String implements fmt.Stringer
func (c *Client) String() string {
	return fmt.Sprintf("{SessionId:%s Token:%s Endpoint:%s}", c.GetSessionId(), c.GetToken(), c.GetEndpoint())
}

// GetSessionId 获取会话Id
func (c *Client) GetSessionId() string {
	return c.sessionId
}

// GetToken 获取token
func (c *Client) GetToken() string {
	return c.options.AuthToken
}

// GetEndpoint 获取服务器地址
func (c *Client) GetEndpoint() string {
	return c.endpoint
}

// GetLocalAddr 获取本地地址
func (c *Client) GetLocalAddr() net.Addr {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.transceiver.Conn.LocalAddr()
}

// GetRemoteAddr 获取对端地址
func (c *Client) GetRemoteAddr() net.Addr {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.transceiver.Conn.RemoteAddr()
}

// SendData 发送数据
func (c *Client) SendData(data []byte) error {
	return c.trans.SendData(data)
}

// SendEvent 发送自定义事件
func (c *Client) SendEvent(event transport.Event[gtp.Msg]) error {
	return transport.Retry{
		Transceiver: &c.transceiver,
		Times:       c.options.IORetryTimes,
	}.Send(c.transceiver.Send(event))
}

// SendDataChan 发送数据的channel
func (c *Client) SendDataChan() chan<- []byte {
	if c.options.SendDataChan == nil {
		c.logger.Panic("send data channel size less equal 0, can't be used")
	}
	return c.options.SendDataChan
}

// RecvDataChan 接收数据的channel
func (c *Client) RecvDataChan() <-chan []byte {
	if c.options.RecvDataChan == nil {
		c.logger.Panic("receive data channel size less equal 0, can't be used")
	}
	return c.options.RecvDataChan
}

// SendEventChan 发送自定义事件的channel
func (c *Client) SendEventChan() chan<- transport.Event[gtp.Msg] {
	if c.options.SendEventChan == nil {
		c.logger.Panic("send event channel size less equal 0, can't be used")
	}
	return c.options.SendEventChan
}

// RecvEventChan 接收自定义事件的channel
func (c *Client) RecvEventChan() <-chan transport.Event[gtp.Msg] {
	if c.options.RecvEventChan == nil {
		c.logger.Panic("receive event channel size less equal 0, can't be used")
	}
	return c.options.RecvEventChan
}

// GetPromise 获取异步编程模型承诺（Promise）
func (c *Client) GetPromise() transport.IPromise {
	return &c.promise
}

// Close 关闭
func (c *Client) Close() {
	c.cancel()
}
