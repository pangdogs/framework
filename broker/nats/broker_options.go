package nats

import (
	"github.com/nats-io/nats.go"
	"net"
	"strings"
)

type WithOption struct{}

type BrokerOptions struct {
	NatsClient    *nats.Conn
	TopicPrefix   string
	QueuePrefix   string
	FastAddresses []string
	FastUsername  string
	FastPassword  string
}

type BrokerOption func(options *BrokerOptions)

func (WithOption) Default() BrokerOption {
	return func(options *BrokerOptions) {
		WithOption{}.NatsClient(nil)(options)
		WithOption{}.TopicPrefix("golaxy.")(options)
		WithOption{}.QueuePrefix("golaxy.")(options)
		WithOption{}.FastAuth("", "")(options)
		WithOption{}.FastAddresses("127.0.0.1:4222")(options)
	}
}

func (WithOption) NatsClient(cli *nats.Conn) BrokerOption {
	return func(o *BrokerOptions) {
		o.NatsClient = cli
	}
}

func (WithOption) TopicPrefix(prefix string) BrokerOption {
	return func(o *BrokerOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
		o.TopicPrefix = prefix
	}
}

func (WithOption) QueuePrefix(prefix string) BrokerOption {
	return func(o *BrokerOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
		o.QueuePrefix = prefix
	}
}

func (WithOption) FastAuth(username, password string) BrokerOption {
	return func(options *BrokerOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

func (WithOption) FastAddresses(addrs ...string) BrokerOption {
	return func(options *BrokerOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(err)
			}
		}
		options.FastAddresses = addrs
	}
}
