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
	"errors"

	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/framework/addins/rpcstack"
)

var (
	// ErrUndeliverable 表示没有处理器可投递当前 RPC。
	ErrUndeliverable = errors.New("rpc: undeliverable")
	// ErrTerminated 表示 RPC 处理器已停止接收调用。
	ErrTerminated = errors.New("rpc: processor terminated")
	// ErrEntityNotFound 表示找不到会话路由所关联的实体。
	ErrEntityNotFound = errors.New("rpc: routing to entity not found")
	// ErrSessionNotFound 表示找不到实体路由所关联的会话。
	ErrSessionNotFound = errors.New("rpc: routing to session not found")
	// ErrGroupNotFound 表示找不到客户端分组。
	ErrGroupNotFound = errors.New("rpc: group not found")
	// ErrDistEntityNotFound 表示分布式实体未注册。
	ErrDistEntityNotFound = errors.New("rpc: distributed entity not found")
	// ErrDistEntityNodeNotFound 表示分布式实体没有匹配的服务节点。
	ErrDistEntityNodeNotFound = errors.New("rpc: distributed entity node not found")
	// ErrIncorrectDestAddress 表示目标地址不符合 RPC 路由格式。
	ErrIncorrectDestAddress = errors.New("rpc: incorrect destination Address")
	// ErrAddInNotFound 表示目标服务或运行时插件不存在。
	ErrAddInNotFound = errors.New("rpc: add-in not found")
	// ErrAddInInactive 表示目标插件未处于运行状态。
	ErrAddInInactive = errors.New("rpc: add-in is inactive")
	// ErrMethodNotFound 表示目标对象未提供指定方法。
	ErrMethodNotFound = errors.New("rpc: method not found")
	// ErrComponentNotFound 表示目标实体没有指定组件。
	ErrComponentNotFound = errors.New("rpc: component not found")
	// ErrMethodParameterCountMismatch 表示调用参数数量与方法签名不匹配。
	ErrMethodParameterCountMismatch = errors.New("rpc: method parameter count mismatch")
	// ErrMethodParameterTypeMismatch 表示调用参数类型与方法签名不匹配。
	ErrMethodParameterTypeMismatch = errors.New("rpc: method parameter type mismatch")
	// ErrAsyncMethodReturnedNil 表示异步 RPC 方法返回了 nil Future。
	ErrAsyncMethodReturnedNil = errors.New("rpc: async method returned nil")
	// ErrPermissionDenied 表示调用路径未通过权限校验。
	ErrPermissionDenied = errors.New("rpc: permission denied")
)

// IDeliverer 选择并投递 RPC 请求或通知。
type IDeliverer interface {
	// Match 报告投递器是否接受当前目标和调用路径。
	Match(svcCtx service.Context, dst string, cc rpcstack.CallChain, cp callpath.CallPath, oneway bool) bool
	// Request 投递需要响应的请求。
	Request(svcCtx service.Context, dst string, cc rpcstack.CallChain, cp callpath.CallPath, args []any) async.Future
	// Notify 投递无需响应的通知。
	Notify(svcCtx service.Context, dst string, cc rpcstack.CallChain, cp callpath.CallPath, args []any) error
}
