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

//go:generate stringer -type SessionState
package gate

import (
	"context"
	"fmt"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/binaryutil"
	"git.golaxy.org/framework/utils/concurrent"
	"net"
	"slices"
	"sync"
)

// SessionState 客户端会话状态
type SessionState int32

const (
	SessionState_Birth     SessionState = iota // 出生
	SessionState_Handshake                     // 与客户端握手中
	SessionState_Confirmed                     // 已确认客户端连接
	SessionState_Active                        // 客户端活跃
	SessionState_Inactive                      // 客户端不活跃，等待重连恢复中
	SessionState_Death                         // 已过期
)

// ISession 会话
type ISession interface {
	context.Context
	fmt.Stringer
	// GetContext 获取服务上下文
	GetContext() service.Context
	// GetId 获取会话Id
	GetId() uid.Id
	// GetUserId 获取用户Id
	GetUserId() string
	// GetToken 获取token
	GetToken() string
	// GetState 获取会话状态
	GetState() SessionState
	// GetLocalAddr 获取本地地址
	GetLocalAddr() net.Addr
	// GetRemoteAddr 获取对端地址
	GetRemoteAddr() net.Addr
	// GetSettings 获取配置
	GetSettings() SessionSettings
	// SendData 发送数据
	SendData(data []byte) error
	// WatchData 监听数据
	WatchData(ctx context.Context, handler SessionRecvDataHandler) IWatcher
	// SendEvent 发送自定义事件
	SendEvent(event transport.IEvent) error
	// WatchEvent 监听自定义事件
	WatchEvent(ctx context.Context, handler SessionRecvEventHandler) IWatcher
	// SendDataChan 发送数据的channel
	SendDataChan() chan<- binaryutil.RecycleBytes
	// RecvDataChan 接收数据的channel
	RecvDataChan() <-chan binaryutil.RecycleBytes
	// SendEventChan 发送自定义事件的channel
	SendEventChan() chan<- transport.IEvent
	// RecvEventChan 接收自定义事件的channel
	RecvEventChan() <-chan transport.IEvent
	// Close 关闭
	Close(err error) <-chan struct{}
	// Closed 已关闭
	Closed() <-chan struct{}
}

type _Session struct {
	context.Context
	sync.Mutex
	terminate       context.CancelCauseFunc
	terminated      chan struct{}
	options         _SessionOptions
	gate            *_Gate
	id              uid.Id
	userId          string
	token           string
	state           SessionState
	transceiver     transport.Transceiver
	eventDispatcher transport.EventDispatcher
	trans           transport.TransProtocol
	ctrl            transport.CtrlProtocol
	renewChan       chan struct{}
	dataWatchers    concurrent.LockedSlice[*_DataWatcher]
	eventWatchers   concurrent.LockedSlice[*_EventWatcher]
}

// String implements fmt.Stringer
func (s *_Session) String() string {
	return fmt.Sprintf(`{"id":%q, "user_id":%q, "token":%q, "state":%d}`, s.GetId(), s.GetUserId(), s.GetToken(), s.GetState())
}

// GetContext 获取服务上下文
func (s *_Session) GetContext() service.Context {
	return s.gate.svcCtx
}

// GetId 获取会话Id
func (s *_Session) GetId() uid.Id {
	return s.id
}

// GetUserId 获取用户Id
func (s *_Session) GetUserId() string {
	return s.userId
}

// GetToken 获取token
func (s *_Session) GetToken() string {
	return s.token
}

// GetState 获取会话状态
func (s *_Session) GetState() SessionState {
	s.Lock()
	defer s.Unlock()
	return s.state
}

// GetLocalAddr 获取本地地址
func (s *_Session) GetLocalAddr() net.Addr {
	s.Lock()
	defer s.Unlock()
	return s.transceiver.Conn.LocalAddr()
}

// GetRemoteAddr 获取对端地址
func (s *_Session) GetRemoteAddr() net.Addr {
	s.Lock()
	defer s.Unlock()
	return s.transceiver.Conn.RemoteAddr()
}

// GetSettings 获取设置
func (s *_Session) GetSettings() SessionSettings {
	s.Lock()
	defer s.Unlock()
	return SessionSettings{
		session:                    s,
		CurrStateChangedHandler:    slices.Clone(s.options.StateChangedHandler),
		CurrSendDataChanSize:       len(s.options.SendEventChan),
		CurrRecvDataChanSize:       len(s.options.RecvDataHandler),
		CurrRecvDataChanRecyclable: s.options.RecvDataChanRecyclable,
		CurrSendEventChanSize:      len(s.options.SendEventChan),
		CurrRecvEventChanSize:      len(s.options.RecvEventChan),
		CurrRecvDataHandler:        slices.Clone(s.options.RecvDataHandler),
		CurrRecvEventHandler:       slices.Clone(s.options.RecvEventHandler),
	}
}

// SendData 发送数据
func (s *_Session) SendData(data []byte) error {
	select {
	case <-s.Done():
		return context.Canceled
	default:
		break
	}
	return s.trans.SendData(data)
}

// WatchData 监听数据
func (s *_Session) WatchData(ctx context.Context, handler SessionRecvDataHandler) IWatcher {
	return s.newDataWatcher(ctx, handler)
}

// SendEvent 发送自定义事件
func (s *_Session) SendEvent(event transport.IEvent) error {
	select {
	case <-s.Done():
		return context.Canceled
	default:
		break
	}
	return transport.Retry{
		Transceiver: &s.transceiver,
		Times:       s.gate.options.IORetryTimes,
	}.Send(s.transceiver.Send(event))
}

// WatchEvent 监听自定义事件
func (s *_Session) WatchEvent(ctx context.Context, handler SessionRecvEventHandler) IWatcher {
	return s.newEventWatcher(ctx, handler)
}

// SendDataChan 发送数据的channel
func (s *_Session) SendDataChan() chan<- binaryutil.RecycleBytes {
	if s.options.SendDataChan == nil {
		log.Panicf(s.gate.svcCtx, "send data channel size less equal 0, can't be used")
	}
	return s.options.SendDataChan
}

// RecvDataChan 接收数据的channel
func (s *_Session) RecvDataChan() <-chan binaryutil.RecycleBytes {
	if s.options.RecvDataChan == nil {
		log.Panicf(s.gate.svcCtx, "receive data channel size less equal 0, can't be used")
	}
	return s.options.RecvDataChan
}

// SendEventChan 发送自定义事件的channel
func (s *_Session) SendEventChan() chan<- transport.IEvent {
	if s.options.SendEventChan == nil {
		log.Panicf(s.gate.svcCtx, "send event channel size less equal 0, can't be used")
	}
	return s.options.SendEventChan
}

// RecvEventChan 接收自定义事件的channel
func (s *_Session) RecvEventChan() <-chan transport.IEvent {
	if s.options.RecvEventChan == nil {
		log.Panicf(s.gate.svcCtx, "receive event channel size less equal 0, can't be used")
	}
	return s.options.RecvEventChan
}

// Close 关闭
func (s *_Session) Close(err error) <-chan struct{} {
	s.terminate(err)
	return s.terminated
}

// Closed 已关闭
func (s *_Session) Closed() <-chan struct{} {
	return s.terminated
}
