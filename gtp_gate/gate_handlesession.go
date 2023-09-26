package gtp_gate

import (
	"fmt"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/logger"
	"net"
	"sync/atomic"
)

// HandleSession 处理会话
func (g *_GtpGate) HandleSession(conn net.Conn) {
	var err error

	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", golaxy.ErrPanicked, panicErr)
		}
		if err != nil {
			logger.Errorf(g.ctx, "listener %q accept client %q, handle session failed, %s", conn.LocalAddr(), conn.RemoteAddr(), err)
			conn.Close()
		}
	}()

	// 网络连接接受器
	acceptor := _Acceptor{
		Gate:    g,
		Options: &g.options,
	}

	// 接受网络连接
	session, err := acceptor.Accept(conn)
	if err != nil {
		return
	}

	logger.Infof(g.ctx, "listener %q accept client %q, handle session success, id: %s, token: %s", conn.LocalAddr(), conn.RemoteAddr(), session.GetId(), session.GetToken())
}

// LoadSession 查询会话
func (g *_GtpGate) LoadSession(sessionId string) (*_GtpSession, bool) {
	v, ok := g.sessionMap.Load(sessionId)
	if !ok {
		return nil, false
	}
	return v.(*_GtpSession), true
}

// StoreSession 存储会话
func (g *_GtpGate) StoreSession(session *_GtpSession) {
	g.sessionMap.Store(session.GetId(), session)
	atomic.AddInt64(&g.sessionCount, 1)
}

// DeleteSession 删除会话
func (g *_GtpGate) DeleteSession(id string) {
	g.sessionMap.Delete(id)
	atomic.AddInt64(&g.sessionCount, -1)
}

// ValidateSession 会话有效性
func (g *_GtpGate) ValidateSession(session *_GtpSession) bool {
	return g.sessionMap.CompareAndSwap(session.GetId(), session, session)
}
