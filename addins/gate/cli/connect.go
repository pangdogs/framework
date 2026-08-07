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

package cli

import (
	"context"
	"fmt"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/option"
)

// Connect 使用 settings 连接 endpoint，完成 GTP 握手后启动客户端循环。
// TCP 模式下 endpoint 为 host:port；WebSocket 模式下为 ws/wss URL。
// ctx 为 nil 时使用 context.Background，取消 ctx 会关闭返回的客户端。
func Connect(ctx context.Context, endpoint string, settings ...option.Setting[ClientOptions]) (*Client, error) {
	connector := _Connector{
		options: option.New(With.Default(), settings...),
	}
	return connector.connect(ctx, endpoint)
}

// Reconnect 为尚未关闭的 client 建立新连接并迁移现有 GTP 会话。
func Reconnect(client *Client) error {
	if client == nil {
		return fmt.Errorf("cli: %w: client is nil", core.ErrArgs)
	}
	connector := _Connector{
		options: client.options,
	}
	return connector.reconnect(client)
}
