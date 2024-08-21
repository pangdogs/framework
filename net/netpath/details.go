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

// Domain 域
type Domain struct {
	Path string // 路径
	Sep  string // 分隔符
}

// IsValid 是否有效
func (d Domain) IsValid() bool {
	return d.Path != "" && d.Sep != ""
}

// Contains 包含路径
func (d Domain) Contains(path string) bool {
	return InDir(d.Sep, path, d.Path)
}

// Equal 等于路径
func (d Domain) Equal(path string) bool {
	return Equal(d.Sep, path, d.Path)
}

// Join 拼接路径
func (d Domain) Join(elems ...string) string {
	return Join(d.Sep, append([]string{d.Path}, elems...)...)
}

// Relative 相对路径
func (d Domain) Relative(path string) (string, bool) {
	if !d.Contains(path) {
		return "", false
	}
	return strings.TrimPrefix(strings.TrimPrefix(path, d.Path), d.Sep), true
}

// NodeDetails 节点地址信息
type NodeDetails struct {
	DomainRoot      Domain // 根域
	DomainBroadcast Domain // 广播地址子域
	DomainBalance   Domain // 负载均衡地址子域
	DomainMulticast Domain // 组播地址子域
	DomainUnicast   Domain // 单播地址子域
}
