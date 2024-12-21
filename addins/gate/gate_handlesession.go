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

package gate

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/log"
	"net"
	"sync/atomic"
)

// handleSession 处理会话
func (g *_Gate) handleSession(conn net.Conn) (*_Session, bool) {
	var err error

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
		if err != nil {
			log.Errorf(g.svcCtx, "listener %q accept remote %q, handle session failed, %s", conn.LocalAddr(), conn.RemoteAddr(), err)
			conn.Close()
		}
	}()

	// 网络连接接受器
	acceptor := _Acceptor{
		gate: g,
	}

	// 接受网络连接
	session, err := acceptor.accept(conn)
	if err != nil {
		return nil, false
	}

	log.Infof(g.svcCtx, "listener %q accept remote %q, handle session success, id:%q, token:%q", conn.LocalAddr(), conn.RemoteAddr(), session.GetId(), session.GetToken())
	return session, true
}

// getSession 查询会话
func (g *_Gate) getSession(sessionId uid.Id) (*_Session, bool) {
	v, ok := g.sessionMap.Load(sessionId)
	if !ok {
		return nil, false
	}
	return v.(*_Session), true
}

// storeSession 存储会话
func (g *_Gate) storeSession(session *_Session) {
	g.sessionMap.Store(session.GetId(), session)
	atomic.AddInt64(&g.sessionCount, 1)
}

// deleteSession 删除会话
func (g *_Gate) deleteSession(sessionId uid.Id) {
	g.sessionMap.Delete(sessionId)
	atomic.AddInt64(&g.sessionCount, -1)
}

// validateSession 会话有效性
func (g *_Gate) validateSession(session *_Session) bool {
	return g.sessionMap.CompareAndSwap(session.GetId(), session, session)
}
