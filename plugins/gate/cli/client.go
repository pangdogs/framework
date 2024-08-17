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
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/binaryutil"
	"git.golaxy.org/framework/utils/concurrent"
	"go.uber.org/zap"
	"net"
	"sync"
)

var (
	ErrReconnectFailed = errors.New("cli: reconnect failed")
	ErrInactiveTimeout = errors.New("cli: inactive timeout")
)

// IWatcher 监听器
type IWatcher interface {
	context.Context
	Terminate() <-chan struct{}
	Terminated() <-chan struct{}
}

// Client 客户端
type Client struct {
	context.Context
	terminate       context.CancelCauseFunc
	terminated      chan struct{}
	wg              sync.WaitGroup
	mutex           sync.Mutex
	options         ClientOptions
	sessionId       uid.Id
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
func (c *Client) GetSessionId() uid.Id {
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

// GetLogger 获取logger
func (c *Client) GetLogger() *zap.SugaredLogger {
	return c.logger
}

// SendData 发送数据
func (c *Client) SendData(data []byte) error {
	return c.trans.SendData(data)
}

// WatchData 监听数据
func (c *Client) WatchData(ctx context.Context, handler RecvDataHandler) IWatcher {
	return c.newDataWatcher(ctx, handler)
}

// SendEvent 发送自定义事件
func (c *Client) SendEvent(event transport.IEvent) error {
	return transport.Retry{
		Transceiver: &c.transceiver,
		Times:       c.options.IORetryTimes,
	}.Send(c.transceiver.Send(event.Interface()))
}

// WatchEvent 监听自定义事件
func (c *Client) WatchEvent(ctx context.Context, handler RecvEventHandler) IWatcher {
	return c.newEventWatcher(ctx, handler)
}

// SendDataChan 发送数据的channel
func (c *Client) SendDataChan() chan<- binaryutil.RecycleBytes {
	if c.options.SendDataChan == nil {
		c.logger.Panic("send data channel size less equal 0, can't be used")
	}
	return c.options.SendDataChan
}

// RecvDataChan 接收数据的channel
func (c *Client) RecvDataChan() <-chan binaryutil.RecycleBytes {
	if c.options.RecvDataChan == nil {
		c.logger.Panic("receive data channel size less equal 0, can't be used")
	}
	return c.options.RecvDataChan
}

// SendEventChan 发送自定义事件的channel
func (c *Client) SendEventChan() chan<- transport.IEvent {
	if c.options.SendEventChan == nil {
		c.logger.Panic("send event channel size less equal 0, can't be used")
	}
	return c.options.SendEventChan
}

// RecvEventChan 接收自定义事件的channel
func (c *Client) RecvEventChan() <-chan transport.IEvent {
	if c.options.RecvEventChan == nil {
		c.logger.Panic("receive event channel size less equal 0, can't be used")
	}
	return c.options.RecvEventChan
}

// Close 关闭
func (c *Client) Close(err error) <-chan struct{} {
	c.terminate(err)
	return c.terminated
}
