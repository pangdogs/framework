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
)

// DeliveryReliability 表示 broker 对单条消息提供的投递保证。
type DeliveryReliability int8

const (
	// DeliveryReliability_AtMostOnce 表示消息最多投递一次，丢失后不会重发。
	DeliveryReliability_AtMostOnce DeliveryReliability = iota
	// DeliveryReliability_AtLeastOnce 表示消息至少投递一次，消费方需要处理重复消息。
	DeliveryReliability_AtLeastOnce
)

// Event 描述一次 broker 消息投递。
type Event struct {
	// Pattern 是创建订阅时使用的话题模式。
	Pattern string
	// Topic 是消息实际发布到的话题。
	Topic string
	// Queue 是订阅所属的队列组；空字符串表示普通订阅。
	Queue string
	// Message 是消息负载，其所有权规则由具体 broker 实现决定。
	Message []byte
	// Ack 向支持确认机制的 broker 确认消息；不支持时返回错误。
	Ack func(ctx context.Context) error
	// Nak 向支持确认机制的 broker 拒绝消息；不支持时返回错误。
	Nak func(ctx context.Context) error
}

type (
	// EventHandler 处理一次 broker 消息投递。
	EventHandler = generic.DelegateVoid1[Event]
)

// IBroker 定义发布、订阅及连接状态刷新所需的消息代理能力。
type IBroker interface {
	// Publish 向 topic 发布 data。
	Publish(ctx context.Context, topic string, data []byte) error
	// SubscribeEvent 订阅 pattern，并在 ctx 取消时关闭返回的事件流。
	// 非空 queue 会创建队列组订阅；autoAck 的支持情况由具体实现决定。
	SubscribeEvent(ctx context.Context, pattern, queue string, autoAck ...bool) (<-chan Event, error)
	// SubscribeHandler 订阅 pattern 并调用 handler；返回的 Future 在订阅结束后完成。
	// 非空 queue 会创建队列组订阅；autoAck 的支持情况由具体实现决定。
	SubscribeHandler(ctx context.Context, pattern, queue string, handler EventHandler, autoAck ...bool) (async.Future, error)
	// Flush 等待当前已缓冲的发布操作发送完成。
	Flush(ctx context.Context) error
	// DeliveryReliability 返回当前实现的投递保证。
	DeliveryReliability() DeliveryReliability
	// MaxPayload 返回单条消息允许的最大负载字节数。
	MaxPayload() int64
	// Separator 返回层级话题的分隔符。
	Separator() string
}
