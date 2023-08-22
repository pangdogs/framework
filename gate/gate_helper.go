package gate

import "kit.golaxy.org/golaxy/service"

// GetSession 查询会话
func GetSession(serviceCtx service.Context, sessionId string) (Session, bool) {
	return Fetch(serviceCtx).GetSession(sessionId)
}

// RangeSessions 遍历所有会话
func RangeSessions(serviceCtx service.Context, fun func(session Session) bool) {
	Fetch(serviceCtx).RangeSessions(fun)
}

// CountSessions 统计所有会话数量
func CountSessions(serviceCtx service.Context) int {
	return Fetch(serviceCtx).CountSessions()
}
