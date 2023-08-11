package gtp

import "kit.golaxy.org/plugins/gate"

// RangeGroups 遍历所有组
func (g *_TcpGate) RangeGroups(fun func(group gate.Group) bool) {
}

// CountGroups 统计所有会话组数量
func (g *_TcpGate) CountGroups() int {
	return 0
}
