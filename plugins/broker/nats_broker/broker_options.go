package nats_broker

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/option"
	"github.com/nats-io/nats.go"
	"net"
	"strings"
)

// Option is a struct used for setting options.
type Option struct{}

// BrokerOptions is a struct that holds various configuration options for the NATS broker.
type BrokerOptions struct {
	NatsClient      *nats.Conn
	TopicPrefix     string
	QueuePrefix     string
	CustomAddresses []string
	CustomUsername  string
	CustomPassword  string
}

// Default sets default values for BrokerOptions.
func (Option) Default() option.Setting[BrokerOptions] {
	return func(options *BrokerOptions) {
		Option{}.NatsClient(nil)(options)
		Option{}.TopicPrefix("")(options)
		Option{}.QueuePrefix("")(options)
		Option{}.CustomAuth("", "")(options)
		Option{}.CustomAddresses("127.0.0.1:4222")(options)
	}
}

// NatsClient sets the NATS client in BrokerOptions.
func (Option) NatsClient(cli *nats.Conn) option.Setting[BrokerOptions] {
	return func(o *BrokerOptions) {
		o.NatsClient = cli
	}
}

// TopicPrefix sets the topic prefix in BrokerOptions.
func (Option) TopicPrefix(prefix string) option.Setting[BrokerOptions] {
	return func(o *BrokerOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
		o.TopicPrefix = prefix
	}
}

// QueuePrefix sets the queue prefix in BrokerOptions.
func (Option) QueuePrefix(prefix string) option.Setting[BrokerOptions] {
	return func(o *BrokerOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
		o.QueuePrefix = prefix
	}
}

// CustomAuth sets the authentication credentials in BrokerOptions. If NatsClient is nil, these credentials are used for authentication.
func (Option) CustomAuth(username, password string) option.Setting[BrokerOptions] {
	return func(options *BrokerOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses sets the addresses in BrokerOptions. If NatsClient is nil, these addresses are used as the connection addresses.
func (Option) CustomAddresses(addrs ...string) option.Setting[BrokerOptions] {
	return func(options *BrokerOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(fmt.Errorf("%w: %w", core.ErrArgs, err))
			}
		}
		options.CustomAddresses = addrs
	}
}
