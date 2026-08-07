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

package rpcpcsr

import (
	"strings"

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/framework/addins/rpcstack"
)

// PermissionValidator 校验调用链是否有权访问调用路径。
type PermissionValidator = generic.Delegate2[rpcstack.CallChain, callpath.CallPath, bool]

// DefaultValidateCliPermission 允许服务间调用，并将客户端 RPC 限制为调用 C_ 前缀的方法。
func DefaultValidateCliPermission(cc rpcstack.CallChain, cp callpath.CallPath) bool {
	if !gate.ClientDetails.DomainRoot.Contains(cc.First().Addr) {
		return true
	}
	return strings.HasPrefix(cp.Method, "C_")
}
