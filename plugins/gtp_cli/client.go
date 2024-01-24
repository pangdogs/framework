package gtp_cli

import (
	"context"
	"fmt"
	"git.golaxy.org/framework/plugins/gtp"
	"git.golaxy.org/framework/plugins/gtp/transport"
	"git.golaxy.org/framework/plugins/util/concurrent"
	"go.uber.org/zap"
	"net"
	"sync"
)

// Watcher 监听器
type Watcher interface {
	context.Context
	Stop() <-chan struct{}
}

// Client 客户端
type Client struct {
	context.Context
	cancel          context.CancelFunc
	closedChan      chan struct{}
	wg              sync.WaitGroup
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
	futures         concurrent.Futures
	dataWatchers    concurrent.LockedSlice[*_DataWatcher]
	eventWatchers   concurrent.LockedSlice[*_EventWatcher]
	logger          *zap.SugaredLogger
}

// String implements fmt.Stringer
func (c *Client) String() string {
	return fmt.Sprintf(`{"session_id":%q, "token":%q, "end_point":%q}`, c.GetSessionId(), c.GetToken(), c.GetEndpoint())
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

// GetFutures 获取异步模型Future控制器
func (c *Client) GetFutures() concurrent.IFutures {
	return &c.futures
}

// SendData 发送数据
func (c *Client) SendData(data []byte) error {
	return c.trans.SendData(data)
}

// WatchData 监听数据
func (c *Client) WatchData(ctx context.Context, handler RecvDataHandler) Watcher {
	return c.newDataWatcher(ctx, handler)
}

// SendEvent 发送自定义事件
func (c *Client) SendEvent(event transport.Event[gtp.MsgReader]) error {
	return transport.Retry{
		Transceiver: &c.transceiver,
		Times:       c.options.IORetryTimes,
	}.Send(c.transceiver.Send(event.Pack()))
}

// WatchEvent 监听自定义事件
func (c *Client) WatchEvent(ctx context.Context, handler RecvEventHandler) Watcher {
	return c.newEventWatcher(ctx, handler)
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
func (c *Client) SendEventChan() chan<- transport.Event[gtp.MsgReader] {
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

// Close 关闭
func (c *Client) Close() <-chan struct{} {
	c.cancel()
	return c.closedChan
}
