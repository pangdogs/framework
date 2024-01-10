package nats_broker

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/option"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/broker"
	"kit.golaxy.org/plugins/log"
	"strings"
	"sync"
)

func newBroker(settings ...option.Setting[BrokerOptions]) broker.Broker {
	return &_Broker{
		options: option.Make(Option{}.Default(), settings...),
	}
}

type _Broker struct {
	ctx     context.Context
	cancel  context.CancelFunc
	servCtx service.Context
	wg      sync.WaitGroup
	options BrokerOptions
	client  *nats.Conn
}

// InitSP 初始化服务插件
func (b *_Broker) InitSP(ctx service.Context) {
	log.Infof(ctx, "init service plugin <%s>:[%s]", plugin.Name, types.AnyFullName(*b))

	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.servCtx = ctx

	if b.options.NatsClient == nil {
		client, err := nats.Connect(strings.Join(b.options.FastAddresses, ","), nats.UserInfo(b.options.FastUsername, b.options.FastPassword), nats.Name(ctx.String()))
		if err != nil {
			log.Panicf(ctx, "connect nats %q failed, %s", b.options.FastAddresses, err)
		}
		b.client = client
	} else {
		b.client = b.options.NatsClient
	}

	if _, err := b.client.RTT(); err != nil {
		log.Panicf(ctx, "rtt nats %q failed, %s", b.client.Servers(), err)
	}
}

// ShutSP 关闭服务插件
func (b *_Broker) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut service plugin <%s>:[%s]", plugin.Name, types.AnyFullName(*b))

	b.cancel()
	b.wg.Wait()

	if b.options.NatsClient == nil {
		if b.client != nil {
			if err := b.client.Drain(); err != nil {
				log.Errorf(ctx, "nats drain failed, %s", err)
			}
		}
	}
}

// Publish the data argument to the given topic. The data argument is left untouched and needs to be correctly interpreted on the receiver.
func (b *_Broker) Publish(ctx context.Context, topic string, data []byte) error {
	if b.options.TopicPrefix != "" {
		topic = b.options.TopicPrefix + topic
	}

	if err := b.client.Publish(topic, data); err != nil {
		return fmt.Errorf("%w: %w", broker.ErrBroker, err)
	}

	return nil
}

// Subscribe will express interest in the given topic pattern. Use option EventHandler to handle message events.
func (b *_Broker) Subscribe(ctx context.Context, pattern string, settings ...option.Setting[broker.SubscriberOptions]) (broker.Subscriber, error) {
	return b.newSubscriber(ctx, _SubscribeMode_Handler, pattern, option.Make(broker.Option{}.Default(), settings...))
}

// SubscribeSync will express interest in the given topic pattern.
func (b *_Broker) SubscribeSync(ctx context.Context, pattern string, settings ...option.Setting[broker.SubscriberOptions]) (broker.SyncSubscriber, error) {
	return b.newSubscriber(ctx, _SubscribeMode_Sync, pattern, option.Make(broker.Option{}.Default(), settings...))
}

// SubscribeChan will express interest in the given topic pattern.
func (b *_Broker) SubscribeChan(ctx context.Context, pattern string, settings ...option.Setting[broker.SubscriberOptions]) (broker.ChanSubscriber, error) {
	return b.newSubscriber(ctx, _SubscribeMode_Chan, pattern, option.Make(broker.Option{}.Default(), settings...))
}

// Flush will perform a round trip to the server and return when it receives the internal reply.
func (b *_Broker) Flush(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := b.client.FlushWithContext(ctx); err != nil {
		return fmt.Errorf("%w: %w", broker.ErrBroker, err)
	}

	return nil
}

// GetDeliveryReliability return message delivery reliability.
func (b *_Broker) GetDeliveryReliability() broker.DeliveryReliability {
	return broker.AtMostOnce
}

// GetMaxPayload return max payload bytes.
func (b *_Broker) GetMaxPayload() int64 {
	return b.client.MaxPayload()
}

// GetSeparator return topic path separator.
func (b *_Broker) GetSeparator() string {
	return "."
}
