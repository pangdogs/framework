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

import "git.golaxy.org/core/service"

// LifecycleInit 由 RPC 插件在处理器启动时调用。
type LifecycleInit interface {
	// Init 初始化处理器。
	Init(svcCtx service.Context)
}

// LifecycleShut 由 RPC 插件在处理器停止时调用。
type LifecycleShut interface {
	// Shut 停止处理器并等待其后台任务退出。
	Shut(svcCtx service.Context)
}
