package gtp_gate

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/framework/plugins/log"
	"net"
	"sync/atomic"
)

// handleSession 处理会话
func (g *_Gate) handleSession(conn net.Conn) {
	var err error

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
		if err != nil {
			log.Errorf(g.servCtx, "listener %q accept client %q, handle session failed, %s", conn.LocalAddr(), conn.RemoteAddr(), err)
			conn.Close()
		}
	}()

	// 网络连接接受器
	acceptor := _Acceptor{
		gate:    g,
		options: &g.options,
	}

	// 接受网络连接
	session, err := acceptor.accept(conn)
	if err != nil {
		return
	}

	log.Infof(g.servCtx, "listener %q accept client %q, handle session success, id: %s, token: %s", conn.LocalAddr(), conn.RemoteAddr(), session.GetId(), session.GetToken())
}

// loadSession 查询会话
func (g *_Gate) loadSession(sessionId string) (*_Session, bool) {
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
func (g *_Gate) deleteSession(id string) {
	g.sessionMap.Delete(id)
	atomic.AddInt64(&g.sessionCount, -1)
}

// validateSession 会话有效性
func (g *_Gate) validateSession(session *_Session) bool {
	return g.sessionMap.CompareAndSwap(session.GetId(), session, session)
}
