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

package dserv

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/netpath"
	"github.com/josharian/intern"
)

// NodeDetails 服务节点地址信息
type NodeDetails struct {
	netpath.NodeDetails
	GlobalBroadcastAddr string // 全局广播地址
	GlobalBalanceAddr   string // 全局负载均衡地址
	BroadcastAddr       string // 服务广播地址
	BalanceAddr         string // 服务负载均衡地址
	LocalAddr           string // 本服务节点地址
}

// MakeBroadcastAddr 创建服务广播地址
func (d *NodeDetails) MakeBroadcastAddr(service string) string {
	return intern.String(d.DomainBroadcast.Join(service))
}

// MakeBalanceAddr 创建服务负载均衡地址
func (d *NodeDetails) MakeBalanceAddr(service string) string {
	return intern.String(d.DomainBalance.Join(service))
}

// MakeNodeAddr 创建服务节点地址
func (d *NodeDetails) MakeNodeAddr(nodeId uid.Id) (string, error) {
	if nodeId.IsNil() {
		return "", fmt.Errorf("%w: nodeId is nil", core.ErrArgs)
	}
	return intern.String(d.DomainNode.Join(nodeId.String())), nil
}
