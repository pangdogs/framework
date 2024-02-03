package cli

import (
	"context"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/option"
)

// Connect 连接服务端
func Connect(ctx context.Context, endpoint string, settings ...option.Setting[ClientOptions]) (*Client, error) {
	connector := _Connector{
		options: option.Make(Option{}.Default(), settings...),
	}
	return connector.connect(ctx, endpoint)
}

// Reonnect 重连服务端
func Reonnect(client *Client) error {
	if client == nil {
		return fmt.Errorf("%w: client is nil", core.ErrArgs)
	}

	connector := _Connector{
		options: client.options,
	}

	return connector.reconnect(client)
}
