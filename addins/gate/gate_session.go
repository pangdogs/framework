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
	"net"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/log"
	"go.uber.org/zap"
)

// establishSession 创建会话
func (g *_Gate) establishSession(conn net.Conn) (*_Session, bool) {
	var err error

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("gate: %w: %w", core.ErrPanicked, panicErr)
		}
		if err != nil {
			log.L(g.svcCtx).Error("failed to establish session",
				zap.String("local", conn.LocalAddr().String()),
				zap.String("remote", conn.RemoteAddr().String()),
				zap.Error(err))
			conn.Close()
		}
	}()

	// 网络连接接受器
	acceptor := _Acceptor{_Gate: g}

	// 接受网络连接
	session, err := acceptor.accept(conn)
	if err != nil {
		log.L(g.svcCtx).Error("failed to establish session",
			zap.String("local", conn.LocalAddr().String()),
			zap.String("remote", conn.RemoteAddr().String()),
			zap.Error(err))
		return nil, false
	}

	log.L(g.svcCtx).Info("session established",
		zap.String("session_id", session.Id().String()),
		zap.String("user_id", session.UserId()),
		zap.String("token", session.Token()),
		zap.Int64("migrations", session.Migrations()),
		zap.String("local", conn.LocalAddr().String()),
		zap.String("remote", conn.RemoteAddr().String()))

	rejected := g.sessionWatcher.Broadcast(session)
	if rejected > 0 {
		addr := session.NetAddr()
		log.L(g.svcCtx).Error("some listeners rejected the session established due to backpressure",
			zap.String("session_id", session.Id().String()),
			zap.String("user_id", session.UserId()),
			zap.String("token", session.Token()),
			zap.Int64("migrations", session.Migrations()),
			zap.String("local", addr.Local.String()),
			zap.String("remote", addr.Remote.String()),
			zap.Int("rejected", rejected))
	}

	return session, true
}

// getSession 查询会话
func (g *_Gate) getSession(id uid.Id) (*_Session, bool) {
	session, ok := g.sessions.Load(id)
	if !ok {
		return nil, false
	}
	return session.(*_Session), true
}

// addSession 添加会话
func (g *_Gate) addSession(session *_Session) bool {
	if _, loaded := g.sessions.LoadOrStore(session.Id(), session); loaded {
		return false
	}
	g.sessionCount.Add(1)
	return true
}

// deleteSession 删除会话
func (g *_Gate) deleteSession(id uid.Id) {
	if _, loaded := g.sessions.LoadAndDelete(id); loaded {
		g.sessionCount.Add(-1)
	}
}

// validateSession 校验会话
func (g *_Gate) validateSession(session *_Session) bool {
	exists, ok := g.sessions.Load(session.Id())
	if !ok {
		return false
	}
	return exists.(*_Session) == session
}
