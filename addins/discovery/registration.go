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

package discovery

import (
	"context"

	"git.golaxy.org/core/utils/async"
)

// IRegistration 控制一次带租约的服务节点注册。
type IRegistration interface {
	// KeepAliveContinuous 持续刷新租约，直到 ctx 取消或保活流结束。
	// 返回的 Future 在保活停止后完成。
	KeepAliveContinuous(ctx context.Context) (async.Future, error)
	// KeepAliveOnce 立即刷新一次租约。
	KeepAliveOnce(ctx context.Context) error
	// Deregister 撤销租约并注销服务节点。
	Deregister(ctx context.Context) error
}
