package gtp_client

import (
	"context"
	"errors"
)

// Connect 连接服务端
func Connect(ctx context.Context, endpoint string, options ...ClientOption) (*Client, error) {
	connector := _Connector{}

	Option{}.Default()(&connector.Options)

	for i := range options {
		options[i](&connector.Options)
	}

	return connector.Connect(ctx, endpoint)
}

// Reonnect 重连服务端
func Reonnect(client *Client) error {
	if client == nil {
		return errors.New("client is nil")
	}

	connector := _Connector{
		Options: client.options,
	}

	return connector.Reconnect(client)
}
