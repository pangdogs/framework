package nats

import (
	"github.com/nats-io/nats.go"
	"net"
	"strings"
)

type Option struct{}

type BrokerOptions struct {
	NatsClient    *nats.Conn
	TopicPrefix   string
	QueuePrefix   string
	FastAddresses []string
	FastUsername  string
	FastPassword  string
}

type BrokerOption func(options *BrokerOptions)

func (Option) Default() BrokerOption {
	return func(options *BrokerOptions) {
		Option{}.NatsClient(nil)(options)
		Option{}.TopicPrefix("golaxy.")(options)
		Option{}.QueuePrefix("golaxy.")(options)
		Option{}.FastAuth("", "")(options)
		Option{}.FastAddresses("127.0.0.1:4222")(options)
	}
}

func (Option) NatsClient(cli *nats.Conn) BrokerOption {
	return func(o *BrokerOptions) {
		o.NatsClient = cli
	}
}

func (Option) TopicPrefix(prefix string) BrokerOption {
	return func(o *BrokerOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
		o.TopicPrefix = prefix
	}
}

func (Option) QueuePrefix(prefix string) BrokerOption {
	return func(o *BrokerOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
		o.QueuePrefix = prefix
	}
}

func (Option) FastAuth(username, password string) BrokerOption {
	return func(options *BrokerOptions) {
		options.FastUsername = username
		options.FastPassword = password
	}
}

func (Option) FastAddresses(addrs ...string) BrokerOption {
	return func(options *BrokerOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(err)
			}
		}
		options.FastAddresses = addrs
	}
}
