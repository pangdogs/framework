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

// NodeDetails 节点地址信息
type NodeDetails struct {
	Domain             string // 主域
	BroadcastSubdomain string // 广播地址子域
	BalanceSubdomain   string // 负载均衡地址子域
	MulticastSubdomain string // 组播地址子域
	NodeSubdomain      string // 单播地址子域
	PathSeparator      string // 地址路径分隔符
}

func (d NodeDetails) InDomain(path string) bool {
	return InDir(d.PathSeparator, path, d.Domain)
}

func (d NodeDetails) EqualDomain(path string) bool {
	return Equal(d.PathSeparator, path, d.Domain)
}

func (d NodeDetails) DomainJoin(elems ...string) string {
	return Join(d.PathSeparator, append([]string{d.Domain}, elems...)...)
}

func (d NodeDetails) InBroadcastSubdomain(path string) bool {
	return InDir(d.PathSeparator, path, d.BroadcastSubdomain)
}

func (d NodeDetails) EqualBroadcastSubdomain(path string) bool {
	return Equal(d.PathSeparator, path, d.BroadcastSubdomain)
}

func (d NodeDetails) BroadcastSubdomainJoin(elems ...string) string {
	return Join(d.PathSeparator, append([]string{d.BroadcastSubdomain}, elems...)...)
}

func (d NodeDetails) InBalanceSubdomain(path string) bool {
	return InDir(d.PathSeparator, path, d.BalanceSubdomain)
}

func (d NodeDetails) EqualBalanceSubdomain(path string) bool {
	return Equal(d.PathSeparator, path, d.BalanceSubdomain)
}

func (d NodeDetails) BalanceSubdomainJoin(elems ...string) string {
	return Join(d.PathSeparator, append([]string{d.BalanceSubdomain}, elems...)...)
}

func (d NodeDetails) InMulticastSubdomain(path string) bool {
	return InDir(d.PathSeparator, path, d.MulticastSubdomain)
}

func (d NodeDetails) EqualMulticastSubdomain(path string) bool {
	return Equal(d.PathSeparator, path, d.MulticastSubdomain)
}

func (d NodeDetails) MulticastSubdomainJoin(elems ...string) string {
	return Join(d.PathSeparator, append([]string{d.MulticastSubdomain}, elems...)...)
}

func (d NodeDetails) InNodeSubdomain(path string) bool {
	return InDir(d.PathSeparator, path, d.NodeSubdomain)
}

func (d NodeDetails) EqualNodeSubdomain(path string) bool {
	return Equal(d.PathSeparator, path, d.NodeSubdomain)
}

func (d NodeDetails) NodeSubdomainJoin(elems ...string) string {
	return Join(d.PathSeparator, append([]string{d.NodeSubdomain}, elems...)...)
}
