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

// IRegistration 服务注册句柄
type IRegistration interface {
	// KeepAliveContinuous 节点持续保活
	KeepAliveContinuous(ctx context.Context) (async.Future, error)
	// KeepAliveOnce 节点保活一次
	KeepAliveOnce(ctx context.Context) error
	// Deregister 注销服务节点
	Deregister(ctx context.Context) error
}
