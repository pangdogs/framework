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
	"net"
	"sync"
	"sync/atomic"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"go.uber.org/zap"
)

// SessionState 客户端会话状态
type SessionState int32

const (
	SessionState_Birth     SessionState = iota // 出生
	SessionState_Confirmed                     // 已确认客户端连接
	SessionState_Active                        // 客户端活跃
	SessionState_Inactive                      // 客户端不活跃，等待重连中
	SessionState_Death                         // 已过期
)

// NetAddr 网络地址
type NetAddr struct {
	Local, Remote net.Addr
}

// ISession 会话
type ISession interface {
	context.Context
	fmt.Stringer
	// Id 获取会话Id
	Id() uid.Id
	// UserId 获取鉴权用户Id
	UserId() string
	// Token 获取鉴权token
	Token() string
	// Extensions 获取鉴权扩展数据
	Extensions() []byte
	// State 获取会话状态
	State() SessionState
	// NetAddr 获取网络地址
	NetAddr() NetAddr
	// Migrations 获取会话连接迁移次数
	Migrations() int64
	// DataIO 获取数据IO
	DataIO() IDataIO
	// EventIO 获取事件IO
	EventIO() IEventIO
	// Close 关闭
	Close(err error) async.Future
	// Closed 已关闭
	Closed() async.Future
}

type _Session struct {
	context.Context
	close           context.CancelCauseFunc
	closed          async.FutureVoid
	gate            *_Gate
	id              uid.Id
	userId          string
	token           string
	extensions      []byte
	state           atomic.Int32
	netAddr         atomic.Pointer[NetAddr]
	transceiver     transport.Transceiver
	eventDispatcher transport.EventDispatcher
	trans           transport.TransProtocol
	ctrl            transport.CtrlProtocol
	migrationMu     sync.Mutex
	migrationChan   chan struct{}
	migrations      atomic.Int64
	io              _SessionIO
	stringerOnce    sync.Once
	stringerCache   string
}

// String implements fmt.Stringer
func (s *_Session) String() string {
	s.stringerOnce.Do(func() {
		s.stringerCache = fmt.Sprintf(`{"id":%q,"user_id":%q}`, s.Id(), s.UserId())
	})
	return s.stringerCache
}

// Id 获取会话Id
func (s *_Session) Id() uid.Id {
	return s.id
}

// UserId 获取鉴权用户Id
func (s *_Session) UserId() string {
	return s.userId
}

// Token 获取鉴权token
func (s *_Session) Token() string {
	return s.token
}

// Extensions 获取鉴权扩展数据
func (s *_Session) Extensions() []byte {
	return s.extensions
}

// State 获取会话状态
func (s *_Session) State() SessionState {
	return SessionState(s.state.Load())
}

// NetAddr 获取网络地址
func (s *_Session) NetAddr() NetAddr {
	return *s.netAddr.Load()
}

// Migrations 获取会话连接迁移次数
func (s *_Session) Migrations() int64 {
	return s.migrations.Load()
}

// DataIO 获取数据IO
func (s *_Session) DataIO() IDataIO {
	return (*_SessionDataIO)(&s.io)
}

// EventIO 获取事件IO
func (s *_Session) EventIO() IEventIO {
	return (*_SessionEventIO)(&s.io)
}

// Close 关闭
func (s *_Session) Close(err error) async.Future {
	s.close(err)
	return s.closed.Out()
}

// Closed 已关闭
func (s *_Session) Closed() async.Future {
	return s.closed.Out()
}

// setState 调整会话状态
func (s *_Session) setState(state SessionState) {
	s.state.Store(int32(state))
}

// handleHeartbeat 处理Heartbeat消息事件
func (s *_Session) handleHeartbeat(event transport.Event[*gtp.MsgHeartbeat]) {
	if event.Flags.Is(gtp.Flag_Ping) {
		log.L(s.gate.svcCtx).Debug("session receive ping", zap.String("session_id", s.Id().String()))
	} else {
		log.L(s.gate.svcCtx).Debug("session receive pong", zap.String("session_id", s.Id().String()))
	}
}
