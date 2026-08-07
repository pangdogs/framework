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

package rpc

import (
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/rpc/rpcpcsr"
)

// RPCOptions 定义 RPC 插件使用的处理器链。
type RPCOptions struct {
	// Processors 按顺序保存处理器；实现 IDeliverer 的处理器也按此顺序参与投递匹配。
	Processors []any
}

// With 提供 RPCOptions 的设置项。
var With _Option

type _Option struct{}

// Default 返回默认设置，默认仅安装服务内 RPC 处理器并启用调用路径压缩。
func (_Option) Default() option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		With.Processors(rpcpcsr.NewServiceProcessor(nil, true))(options)
	}
}

// Processors 替换 RPC 插件使用的处理器链。
func (_Option) Processors(processors ...any) option.Setting[RPCOptions] {
	return func(options *RPCOptions) {
		options.Processors = processors
	}
}
