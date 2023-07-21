package tcp

import (
	"kit.golaxy.org/plugins/gate"
	"sync/atomic"
)

// RangeSessions 遍历所有会话
func (g *_TcpGate) RangeSessions(fun func(session gate.Session) bool) {
	if fun == nil {
		return
	}
	g.sessionMap.Range(func(k, v any) bool {
		return fun(v.(gate.Session))
	})
}

// CountSessions 统计所有会话数量
func (g *_TcpGate) CountSessions() int {
	return int(atomic.LoadInt64(&g.sessionCount))
}
