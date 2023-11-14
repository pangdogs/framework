package gtp_client

import (
	"context"
	"fmt"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/util/option"
)

// Connect 连接服务端
func Connect(ctx context.Context, endpoint string, settings ...option.Setting[ClientOptions]) (*Client, error) {
	connector := _Connector{
		options: option.Make[](Option{}.Default(), settings...),
	}
	return connector.connect(ctx, endpoint)
}

// Reonnect 重连服务端
func Reonnect(client *Client) error {
	if client == nil {
		return fmt.Errorf("%w: client is nil", golaxy.ErrArgs)
	}

	connector := _Connector{
		options: client.options,
	}

	return connector.reconnect(client)
}
