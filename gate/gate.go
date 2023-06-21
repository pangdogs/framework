package gate

// Gate 网关
type Gate interface {
	SessionManager
	GroupManager
	// Broadcast 广播数据
	Broadcast(data []byte) error
	// Multicast 组播数据
	Multicast(groupId string, data []byte) error
	// Unicast 单播数据
	Unicast(sessionId string, data []byte) error
}

// SessionManager 会话管理器
type SessionManager interface {
	// RangeSessions 遍历所有会话
	RangeSessions(fun func(session Session) bool)
	// CountSessions 统计所有会话数量
	CountSessions() int

	//HandleNewSession()
	//HandleDestroySession()
}

// GroupManager 会话组管理器
type GroupManager interface {
	// RangeGroups 遍历所有组
	RangeGroups(fun func(group Group) bool)
	// CountGroups 统计所有会话组数量
	CountGroups() int

	//HandleNewSession()
	//HandleDestroySession()
}
