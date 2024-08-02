/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package nats_broker

import (
	"context"
	"fmt"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/broker"
	"git.golaxy.org/framework/plugins/log"
	"github.com/nats-io/nats.go"
	"strings"
	"sync"
)

func newBroker(settings ...option.Setting[BrokerOptions]) broker.IBroker {
	return &_Broker{
		options: option.Make(With.Default(), settings...),
	}
}

type _Broker struct {
	servCtx   service.Context
	ctx       context.Context
	terminate context.CancelFunc
	wg        sync.WaitGroup
	options   BrokerOptions
	client    *nats.Conn
}

// InitSP 初始化服务插件
func (b *_Broker) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	b.servCtx = ctx
	b.ctx, b.terminate = context.WithCancel(context.Background())

	if b.options.NatsClient == nil {
		client, err := nats.Connect(strings.Join(b.options.CustomAddresses, ","), nats.UserInfo(b.options.CustomUsername, b.options.CustomPassword), nats.Name(ctx.String()))
		if err != nil {
			log.Panicf(ctx, "connect nats %q failed, %s", b.options.CustomAddresses, err)
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
	log.Infof(ctx, "shut plugin %q", self.Name)

	b.terminate()
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
		return fmt.Errorf("broker: %w", err)
	}

	return nil
}

// Subscribe will express interest in the given topic pattern. Use option EventHandler to handle message events.
func (b *_Broker) Subscribe(ctx context.Context, pattern string, settings ...option.Setting[broker.SubscriberOptions]) (broker.ISubscriber, error) {
	return b.newSubscriber(ctx, _SubscribeMode_Handler, pattern, option.Make(broker.With.Default(), settings...))
}

// SubscribeSync will express interest in the given topic pattern.
func (b *_Broker) SubscribeSync(ctx context.Context, pattern string, settings ...option.Setting[broker.SubscriberOptions]) (broker.ISyncSubscriber, error) {
	return b.newSubscriber(ctx, _SubscribeMode_Sync, pattern, option.Make(broker.With.Default(), settings...))
}

// SubscribeChan will express interest in the given topic pattern.
func (b *_Broker) SubscribeChan(ctx context.Context, pattern string, settings ...option.Setting[broker.SubscriberOptions]) (broker.IChanSubscriber, error) {
	return b.newSubscriber(ctx, _SubscribeMode_Chan, pattern, option.Make(broker.With.Default(), settings...))
}

// Flush will perform a round trip to the server and return when it receives the internal reply.
func (b *_Broker) Flush(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := b.client.FlushWithContext(ctx); err != nil {
		return fmt.Errorf("broker: %w", err)
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
