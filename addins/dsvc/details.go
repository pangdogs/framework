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

package dsvc

import (
	"fmt"
	"unique"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/netpath"
)

// NodeDetails 汇总当前节点在分布式服务地址空间中的广播、负载均衡及单播地址。
type NodeDetails struct {
	netpath.NodeDetails
	GlobalBroadcastAddr string `json:"global_broadcast_addr"` // GlobalBroadcastAddr 面向所有服务节点广播。
	GlobalBalanceAddr   string `json:"global_balance_addr"`   // GlobalBalanceAddr 在所有服务节点间负载均衡。
	BroadcastAddr       string `json:"broadcast_addr"`        // BroadcastAddr 面向同名服务的所有节点广播。
	BalanceAddr         string `json:"balance_addr"`          // BalanceAddr 在同名服务节点间负载均衡。
	LocalAddr           string `json:"local_addr"`            // LocalAddr 唯一寻址当前服务节点。
}

// MakeBroadcastAddr 返回指定逻辑服务的广播地址。
func (d *NodeDetails) MakeBroadcastAddr(service string) string {
	return unique.Make(d.DomainBroadcast.Join(service)).Value()
}

// MakeBalanceAddr 返回指定逻辑服务的负载均衡地址。
func (d *NodeDetails) MakeBalanceAddr(service string) string {
	return unique.Make(d.DomainBalance.Join(service)).Value()
}

// MakeNodeAddr 返回指定节点 ID 的单播地址；nodeID 为空时返回错误。
func (d *NodeDetails) MakeNodeAddr(nodeID uid.ID) (string, error) {
	if nodeID.IsNil() {
		return "", fmt.Errorf("dsvc: %w: nodeID is nil", core.ErrArgs)
	}
	return unique.Make(d.DomainUnicast.Join(nodeID.String())).Value(), nil
}
