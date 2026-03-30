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

package broker_nats

import (
	"context"
	"fmt"
	"strings"

	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/broker"
	"git.golaxy.org/framework/addins/log"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func newNatsBroker(settings ...option.Setting[NatsBrokerOptions]) broker.IBroker {
	return &_NatsBroker{
		options: option.New(With.Default(), settings...),
	}
}

type _NatsBroker struct {
	svcCtx    service.Context
	ctx       context.Context
	terminate context.CancelFunc
	barrier   generic.Barrier
	options   NatsBrokerOptions
	client    *nats.Conn
}

// Init 初始化插件
func (b *_NatsBroker) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	b.svcCtx = svcCtx
	b.ctx, b.terminate = context.WithCancel(context.Background())

	if b.options.NatsClient == nil {
		client, err := nats.Connect(strings.Join(b.options.CustomAddresses, ","), nats.UserInfo(b.options.CustomUsername, b.options.CustomPassword), nats.Name(svcCtx.String()))
		if err != nil {
			log.L(svcCtx).Panic("connect nats failed",
				zap.Strings("addresses", b.options.CustomAddresses),
				zap.String("username", b.options.CustomUsername),
				zap.String("password", b.options.CustomPassword),
				zap.Error(err))
		}
		b.client = client
	} else {
		b.client = b.options.NatsClient
	}

	if _, err := b.client.RTT(); err != nil {
		log.L(svcCtx).Panic("rtt nats failed", zap.Strings("servers", b.client.Servers()), zap.Error(err))
	}
}

// Shut 关闭插件
func (b *_NatsBroker) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))

	b.terminate()
	b.barrier.Close()
	b.barrier.Wait()

	if b.options.NatsClient == nil {
		if b.client != nil {
			if err := b.client.Drain(); err != nil {
				log.L(svcCtx).Error("drain nats failed", zap.Strings("servers", b.client.Servers()), zap.Error(err))
			}
		}
	}
}

// Publish 发布
func (b *_NatsBroker) Publish(ctx context.Context, topic string, data []byte) error {
	if b.options.TopicPrefix != "" {
		topic = b.options.TopicPrefix + topic
	}

	if err := b.client.Publish(topic, data); err != nil {
		log.L(b.svcCtx).Error("publish topic failed", zap.String("topic", topic), zap.Error(err))
		return fmt.Errorf("broker: %w", err)
	}

	return nil
}

// SubscribeEvent 订阅消息事件流
func (b *_NatsBroker) SubscribeEvent(ctx context.Context, pattern, queue string, _ ...bool) (<-chan broker.Event, error) {
	eventChan, _, err := b.addSubscriber(ctx, pattern, queue, nil)
	if err != nil {
		return nil, err
	}
	return eventChan, nil
}

// SubscribeHandler 订阅消息事件回调
func (b *_NatsBroker) SubscribeHandler(ctx context.Context, pattern, queue string, handler broker.EventHandler, _ ...bool) (async.Future, error) {
	if handler == nil {
		return async.Future{}, fmt.Errorf("broker: %w: handler is nil", core.ErrArgs)
	}
	_, unsubscribed, err := b.addSubscriber(ctx, pattern, queue, handler)
	if err != nil {
		return async.Future{}, err
	}
	return unsubscribed, nil
}

// Flush 刷新
func (b *_NatsBroker) Flush(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := b.client.FlushWithContext(ctx); err != nil {
		log.L(b.svcCtx).Error("flush failed", zap.Error(err))
		return fmt.Errorf("broker: %w", err)
	}

	return nil
}

// DeliveryReliability 获取消息投递模式
func (b *_NatsBroker) DeliveryReliability() broker.DeliveryReliability {
	return broker.DeliveryReliability_AtMostOnce
}

// MaxPayload 获取最大消息长度
func (b *_NatsBroker) MaxPayload() int64 {
	return b.client.MaxPayload()
}

// Separator 获取地址分隔符
func (b *_NatsBroker) Separator() string {
	return "."
}
