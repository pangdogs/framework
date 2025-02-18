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
	ErrUndeliverable                = errors.New("rpc: undeliverable")                     // 无法投递
	ErrTerminated                   = errors.New("rpc: processor terminated")              // 已终止处理
	ErrEntityNotFound               = errors.New("rpc: routing to entity not found")       // 找不到路由会话映射的实体
	ErrSessionNotFound              = errors.New("rpc: routing to session not found")      // 找不到路由实体映射的会话
	ErrGroupNotFound                = errors.New("rpc: group not found")                   // 找不到分组
	ErrGroupChanIsFull              = errors.New("rpc: group send data channel is full")   // 分组发送数据的channel已满
	ErrDistEntityNotFound           = errors.New("rpc: distributed entity not found")      // 找不到分布式实体
	ErrDistEntityNodeNotFound       = errors.New("rpc: distributed entity node not found") // 找不到分布式实体的服务节点
	ErrIncorrectDestAddress         = errors.New("rpc: incorrect destination Address")     // 错误的目的地址
	ErrAddInNotFound                = errors.New("rpc: addIn not found")                   // 找不到插件
	ErrAddInInactive                = errors.New("rpc: addIn is inactive")                 // 插件未激活
	ErrMethodNotFound               = errors.New("rpc: method not found")                  // 找不到方法
	ErrComponentNotFound            = errors.New("rpc: component not found")               // 找不到组件
	ErrMethodParameterCountMismatch = errors.New("rpc: method parameter count mismatch")   // 方法参数数量不匹配
	ErrMethodParameterTypeMismatch  = errors.New("rpc: method parameter type mismatch")    // 方法参数类型不匹配
	ErrAsyncMethodReturnedNil       = errors.New("rpc: async method returned nil")         // 异步方法返回值为nil
	ErrPermissionDenied             = errors.New("rpc: permission denied")                 // 权限不足
)

// IDeliverer RPC投递器接口
type IDeliverer interface {
	// Match 是否匹配
	Match(svcCtx service.Context, dst string, cc rpcstack.CallChain, cp callpath.CallPath, oneway bool) bool
	// Request 请求
	Request(svcCtx service.Context, dst string, cc rpcstack.CallChain, cp callpath.CallPath, args []any) async.AsyncRet
	// Notify 通知
	Notify(svcCtx service.Context, dst string, cc rpcstack.CallChain, cp callpath.CallPath, args []any) error
}
