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

//go:generate stringer -type DeliveryReliability
package broker

import (
	"context"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
)

// DeliveryReliability 消息投递模式
type DeliveryReliability int8

const (
	DeliveryReliability_AtMostOnce  DeliveryReliability = iota // 最多一次
	DeliveryReliability_AtLeastOnce                            // 最少一次
)

// Event 消息事件
type Event struct {
	// Pattern 订阅话题模式
	Pattern string
	// Topic 订阅话题
	Topic string
	// Queue 订阅队列组
	Queue string
	// Message 消息数据
	Message []byte
	// Ack 确认
	Ack func(ctx context.Context) error
	// Nak 拒绝确认
	Nak func(ctx context.Context) error
}

type (
	// EventHandler 消息事件处理器
	EventHandler = generic.DelegateVoid1[Event]
)

// IBroker 消息中间件接口
type IBroker interface {
	// Publish 发布
	Publish(ctx context.Context, topic string, data []byte) error
	// SubscribeEvent 订阅消息事件流
	SubscribeEvent(ctx context.Context, pattern string, settings ...option.Setting[SubscribeOptions]) (<-chan Event, error)
	// SubscribeHandler 订阅消息事件回调
	SubscribeHandler(ctx context.Context, pattern string, handler EventHandler, settings ...option.Setting[SubscribeOptions]) (async.Future, error)
	// Flush 刷新
	Flush(ctx context.Context) error
	// DeliveryReliability 获取消息投递模式
	DeliveryReliability() DeliveryReliability
	// MaxPayload 获取最大消息长度
	MaxPayload() int64
	// Separator 获取地址分隔符
	Separator() string
}
