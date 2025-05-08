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

package broker

import (
	"context"
	"git.golaxy.org/core/utils/option"
)

// DeliveryReliability Message delivery reliability.
type DeliveryReliability int32

const (
	AtMostOnce      DeliveryReliability = iota // At most once
	AtLeastOnce                                // At last once
	ExactlyOnce                                // Exactly once
	EffectivelyOnce                            // Effectively once
)

// IBroker is an interface used for asynchronous messaging.
type IBroker interface {
	// Publish the data argument to the given topic. The data argument is left untouched and needs to be correctly interpreted on the receiver.
	Publish(ctx context.Context, topic string, data []byte) error
	// Subscribe will express interest in the given topic pattern. Use option EventHandler to handle message events.
	Subscribe(ctx context.Context, pattern string, settings ...option.Setting[SubscriberOptions]) (ISubscriber, error)
	// Subscribef will express interest in the given topic pattern with a formatted string. Use option EventHandler to handle message events.
	Subscribef(ctx context.Context, format string, args ...any) func(settings ...option.Setting[SubscriberOptions]) (ISubscriber, error)
	// Subscribep will express interest in the given topic pattern with elements. Use option EventHandler to handle message events.
	Subscribep(ctx context.Context, elems ...string) func(settings ...option.Setting[SubscriberOptions]) (ISubscriber, error)
	// Flush will perform a round trip to the server and return when it receives the internal reply.
	Flush(ctx context.Context) error
	// GetDeliveryReliability return message delivery reliability.
	GetDeliveryReliability() DeliveryReliability
	// GetMaxPayload return max payload bytes.
	GetMaxPayload() int64
	// GetSeparator return topic path separator.
	GetSeparator() string
}
