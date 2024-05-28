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
		return fmt.Errorf("%w: client is nil", core.ErrArgs)
	}

	connector := _Connector{
		options: client.options,
	}

	return connector.reconnect(client)
}
