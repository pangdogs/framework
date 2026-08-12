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

// SessionState 表示服务端会话从建立到过期的当前阶段。
type SessionState int32

const (
	// SessionState_Birth 表示会话已创建但握手尚未确认。
	SessionState_Birth SessionState = iota
	// SessionState_Confirmed 表示客户端身份与连接已经确认。
	SessionState_Confirmed
	// SessionState_Active 表示会话连接当前可用。
	SessionState_Active
	// SessionState_Inactive 表示连接中断，正在等待迁移重连。
	SessionState_Inactive
	// SessionState_Death 表示会话已关闭并从 gate 删除。
	SessionState_Death
)

// NetAddr 保存会话当前连接的本地与对端地址快照。
type NetAddr struct {
	Local, Remote net.Addr // Local 是服务端地址，Remote 是客户端地址。
}

// ISession 表示可迁移连接的已鉴权客户端会话。
// 会话上下文在 Close 时取消；查询和 I/O 门面可被多个 goroutine 使用。
type ISession interface {
	context.Context
	fmt.Stringer
	// Id 返回服务端分配的会话 ID。
	Id() uid.Id
	// UserId 返回客户端提交的鉴权用户 ID。
	UserId() string
	// Token 返回客户端提交的鉴权令牌。
	Token() string
	// Extensions 返回客户端提交的鉴权扩展数据；调用方不得修改。
	Extensions() []byte
	// State 返回当前会话状态。
	State() SessionState
	// NetAddr 返回当前连接地址快照；连接迁移后会变化。
	NetAddr() NetAddr
	// Migrations 返回连接成功迁移的累计次数。
	Migrations() int64
	// DataIO 返回原始负载 I/O 门面。
	DataIO() IDataIO
	// EventIO 返回 GTP 事件 I/O 门面。
	EventIO() IEventIO
	// Close 请求以 err 为原因关闭会话，并返回关闭完成信号。
	Close(err error) async.Signal
	// Closed 返回会话关闭完成信号。
	Closed() async.Signal
}

type _Session struct {
	context.Context
	close           context.CancelCauseFunc
	closed          async.Completer
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

// String 返回包含会话 ID 和鉴权用户 ID 的缓存 JSON 文本。
func (s *_Session) String() string {
	s.stringerOnce.Do(func() {
		s.stringerCache = fmt.Sprintf(`{"id":%q,"user_id":%q}`, s.Id(), s.UserId())
	})
	return s.stringerCache
}

// Id 返回服务端分配的会话 ID。
func (s *_Session) Id() uid.Id {
	return s.id
}

// UserId 返回客户端提交的鉴权用户 ID。
func (s *_Session) UserId() string {
	return s.userId
}

// Token 返回客户端提交的鉴权令牌。
func (s *_Session) Token() string {
	return s.token
}

// Extensions 返回客户端提交的鉴权扩展数据；调用方不得修改返回切片。
func (s *_Session) Extensions() []byte {
	return s.extensions
}

// State 返回当前会话状态。
func (s *_Session) State() SessionState {
	return SessionState(s.state.Load())
}

// NetAddr 返回当前连接地址快照；连接迁移后会变化。
func (s *_Session) NetAddr() NetAddr {
	return *s.netAddr.Load()
}

// Migrations 返回连接成功迁移的累计次数。
func (s *_Session) Migrations() int64 {
	return s.migrations.Load()
}

// DataIO 返回原始负载 I/O 门面。
func (s *_Session) DataIO() IDataIO {
	return (*_SessionDataIO)(&s.io)
}

// EventIO 返回 GTP 事件 I/O 门面。
func (s *_Session) EventIO() IEventIO {
	return (*_SessionEventIO)(&s.io)
}

// Close 请求以 err 为原因关闭会话，并返回关闭完成信号。
func (s *_Session) Close(err error) async.Signal {
	s.close(err)
	return s.closed.Signal()
}

// Closed 返回会话关闭完成信号。
func (s *_Session) Closed() async.Signal {
	return s.closed.Signal()
}

// setState 原子更新会话状态。
func (s *_Session) setState(state SessionState) {
	s.state.Store(int32(state))
}

// handleHeartbeat 记录收到的 Ping 或 Pong 心跳事件。
func (s *_Session) handleHeartbeat(event transport.Event[*gtp.MsgHeartbeat]) {
	if event.Flags.Is(gtp.Flag_Ping) {
		log.L(s.gate.svcCtx).Debug("session receive ping", zap.String("session_id", s.Id().String()))
	} else {
		log.L(s.gate.svcCtx).Debug("session receive pong", zap.String("session_id", s.Id().String()))
	}
}
