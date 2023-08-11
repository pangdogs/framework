package gtp

import "kit.golaxy.org/plugins/gate"

// RangeGroups 遍历所有组
func (g *_GtpGate) RangeGroups(fun func(group gate.Group) bool) {
}

// CountGroups 统计所有会话组数量
func (g *_GtpGate) CountGroups() int {
	return 0
}
