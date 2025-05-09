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

// Connect 连接服务端
func Connect(ctx context.Context, endpoint string, settings ...option.Setting[ClientOptions]) (*Client, error) {
	connector := _Connector{
		options: option.Make(With.Default(), settings...),
	}
	return connector.connect(ctx, endpoint)
}

// Reconnect 重连服务端
func Reconnect(client *Client) error {
	if client == nil {
		return fmt.Errorf("cli: %w: client is nil", core.ErrArgs)
	}

	connector := _Connector{
		options: client.options,
	}

	return connector.reconnect(client)
}
