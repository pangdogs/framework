package nats_broker

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/util/option"
	"net"
	"strings"
)

// Option is a struct used for setting options.
type Option struct{}

// BrokerOptions is a struct that holds various configuration options for the NATS broker.
type BrokerOptions struct {
	NatsClient    *nats.Conn
	TopicPrefix   string
	QueuePrefix   string
	FastAddresses []string
	FastUsername  string
	FastPassword  string
}

// Default sets default values for BrokerOptions.
func (Option) Default() option.Setting[BrokerOptions] {
	return func(options *BrokerOptions) {
		Option{}.NatsClient(nil)(options)
		Option{}.TopicPrefix("")(options)
		Option{}.QueuePrefix("")(options)
		Option{}.FastAuth("", "")(options)
		Option{}.FastAddresses("127.0.0.1:4222")(options)
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

// FastAuth sets the authentication credentials in BrokerOptions. If NatsClient is nil, these credentials are used for authentication.
func (Option) FastAuth(username, password string) option.Setting[BrokerOptions] {
	return func(options *BrokerOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

// FastAddresses sets the addresses in BrokerOptions. If NatsClient is nil, these addresses are used as the connection addresses.
func (Option) FastAddresses(addrs ...string) option.Setting[BrokerOptions] {
	return func(options *BrokerOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(fmt.Errorf("%w: %w", golaxy.ErrArgs, err))
			}
		}
		options.FastAddresses = addrs
	}
}
