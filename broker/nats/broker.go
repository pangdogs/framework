package nats

import (
	"context"
	"github.com/nats-io/nats.go"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/broker"
	"kit.golaxy.org/plugins/logger"
	"strings"
)

func newNatsBroker(options ...BrokerOption) broker.Broker {
	opts := BrokerOptions{}
	WithOption{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_NatsBroker{
		options: opts,
	}
}

type _NatsBroker struct {
	options BrokerOptions
	ctx     service.Context
	client  *nats.Conn
}

// InitSP 初始化服务插件
func (b *_NatsBroker) InitSP(ctx service.Context) {
	logger.Infof(ctx, "init service plugin %q with %q", definePlugin.Name, util.TypeOfAnyFullName(*b))

	b.ctx = ctx

	if b.options.NatsClient == nil {
		client, err := nats.Connect(strings.Join(b.options.FastAddresses, ","),
			nats.UserInfo(b.options.FastUsername, b.options.FastPassword),
			nats.Name(ctx.String()))
		if err != nil {
			logger.Panicf(ctx, "connect nats %q failed, %s", b.options.FastAddresses, err)
		}
		b.client = client
	} else {
		b.client = b.options.NatsClient
	}

	if _, err := b.client.RTT(); err != nil {
		logger.Panicf(ctx, "rtt nats %q failed, %s", b.client.Servers(), err)
	}
}

// ShutSP 关闭服务插件
func (b *_NatsBroker) ShutSP(ctx service.Context) {
	logger.Infof(ctx, "shut service plugin %q", definePlugin.Name)

	if b.options.NatsClient == nil {
		if b.client != nil {
			if err := b.client.Drain(); err != nil {
				logger.Errorf(ctx, "nats drain failed, %s", err)
			}
		}
	}
}

// Publish the data argument to the given topic. The data argument is left untouched and needs to be correctly interpreted on the receiver.
func (b *_NatsBroker) Publish(ctx context.Context, topic string, data []byte) error {
	return b.client.Publish(topic, data)
}

// Subscribe will express interest in the given topic pattern.
func (b *_NatsBroker) Subscribe(ctx context.Context, pattern string, options ...broker.SubscriberOption) (broker.Subscriber, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	opts := broker.SubscriberOptions{}
	broker.WithOption{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return newNatsSubscriber(ctx, b, pattern, opts)
}

// Flush will perform a round trip to the server and return when it receives the internal reply.
func (b *_NatsBroker) Flush(ctx context.Context) error {
	return b.client.FlushWithContext(ctx)
}

// MaxPayload return max payload bytes.
func (b *_NatsBroker) MaxPayload() int64 {
	return b.client.MaxPayload()
}
