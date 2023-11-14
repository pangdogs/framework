package gtp_gate

import (
	"kit.golaxy.org/golaxy/service"
)

// GetSession 查询会话
func GetSession(servCtx service.Context, sessionId string) (Session, bool) {
	return Using(servCtx).GetSession(sessionId)
}

// RangeSessions 遍历所有会话
func RangeSessions(servCtx service.Context, fun func(session Session) bool) {
	Using(servCtx).RangeSessions(fun)
}

// CountSessions 统计所有会话数量
func CountSessions(servCtx service.Context) int {
	return Using(servCtx).CountSessions()
}
