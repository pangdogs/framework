package gate

import "io"

// Group 客户端会话组
type Group interface {
	io.Writer
	io.ReaderFrom
	// GetId 获取会话组Id
	GetId() string
	// RangeSessions 遍历组内所有会话
	RangeSessions(fun func(session Session) bool)
	// CountSessions 统计组内所有会话数量
	CountSessions() int
}
