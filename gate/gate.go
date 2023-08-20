package gate

// Gate 网关
type Gate interface {
	// GetSession 查询会话
	GetSession(sessionId string) (Session, bool)
	// RangeSessions 遍历所有会话
	RangeSessions(fun func(session Session) bool)
	// CountSessions 统计所有会话数量
	CountSessions() int
}
