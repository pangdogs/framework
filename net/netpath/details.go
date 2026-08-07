/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package netpath

import (
	"strings"
)

// Domain 定义以固定分隔符组织的逻辑地址域。
type Domain struct {
	Path string `json:"path"` // 域路径。
	Sep  string `json:"sep"`  // 路径段分隔符。
}

// IsValid 报告域路径和分隔符是否均非空。
func (d Domain) IsValid() bool {
	return d.Path != "" && d.Sep != ""
}

// Contains 报告 path 是否为当前域的严格子路径；与域本身相等时返回 false。
func (d Domain) Contains(path string) bool {
	return InDir(d.Sep, path, d.Path)
}

// Equal 报告 path 是否与当前域相等；双方末尾的一个分隔符会被忽略。
func (d Domain) Equal(path string) bool {
	return Equal(d.Sep, path, d.Path)
}

// Join 将域路径作为首段，与 elems 按 Sep 原样拼接，不清理空段或重复分隔符。
func (d Domain) Join(elems ...string) string {
	return Join(d.Sep, append([]string{d.Path}, elems...)...)
}

// Relative 返回严格子路径 path 相对于当前域的部分；path 不在域内时返回 false。
func (d Domain) Relative(path string) (string, bool) {
	if !d.Contains(path) {
		return "", false
	}
	return strings.TrimPrefix(strings.TrimPrefix(path, d.Path), d.Sep), true
}

// NodeDetails 描述节点消息地址空间使用的各类逻辑域。
type NodeDetails struct {
	DomainRoot      Domain `json:"domain_root"`      // 节点地址空间的根域。
	DomainBroadcast Domain `json:"domain_broadcast"` // 广播地址域。
	DomainBalance   Domain `json:"domain_balance"`   // 负载均衡地址域。
	DomainMulticast Domain `json:"domain_multicast"` // 组播地址域。
	DomainUnicast   Domain `json:"domain_unicast"`   // 单播地址域。
}
