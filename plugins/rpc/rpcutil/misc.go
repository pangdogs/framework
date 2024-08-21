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

package rpcutil

import (
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/plugins/rpcstack"
	"strings"
)

const (
	NoPlugin = "" // 不使用插件
	NoComp   = "" // 不使用组件
)

// CliRPCPermValidator 默认的客户端RPC请求权限验证器，强制客户端只能RPC调用前缀为C_的函数
func CliRPCPermValidator(cc rpcstack.CallChain, cp callpath.CallPath) bool {
	if !gate.CliDetails.DomainRoot.Contains(cc.First().Addr) {
		return true
	}
	return strings.HasPrefix(cp.Method, "C_")
}
