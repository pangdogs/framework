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
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
)

type (
	// EventHandler is used to process messages via a subscription of a topic. The handler is passed a publication interface which contains the
	// message and optional Ack method to acknowledge receipt of the message.
	EventHandler = generic.Delegate1[Event, error]
	// UnsubscribedCB Unsubscribed callback method.
	UnsubscribedCB = generic.Action1[ISubscriber]
)

// SubscriberOptions represents the options for subscribe topic.
type SubscriberOptions struct {
	// AutoAck defaults to true. When a handler returns with a nil error the message is acked.
	AutoAck bool
	// Queue subscribers with the same queue name will create a shared subscription where each receives a subset of messages.
	Queue string
	// EventChanSize specifies the size of the event channel used for received synchronously event.
	EventChanSize int
	// EventHandler is the function that will be called to handle the received events.
	EventHandler EventHandler
	// UnsubscribedCB Unsubscribed callback method.
	UnsubscribedCB UnsubscribedCB
}

var With _Option

type _Option struct{}

// Default sets the default options for subscribe topic.
func (_Option) Default() option.Setting[SubscriberOptions] {
	return func(options *SubscriberOptions) {
		With.AutoAck(true)(options)
		With.Queue("")(options)
		With.EventChanSize(0)(options)
		With.EventHandler(nil)(options)
		With.UnsubscribedCB(nil)(options)
	}
}

// AutoAck defaults to true. When a handler returns with a nil error the message is acked.
func (_Option) AutoAck(b bool) option.Setting[SubscriberOptions] {
	return func(o *SubscriberOptions) {
		o.AutoAck = b
	}
}

// Queue subscribers with the same queue name will create a shared subscription where each
// receives a subset of messages.
func (_Option) Queue(queue string) option.Setting[SubscriberOptions] {
	return func(o *SubscriberOptions) {
		o.Queue = queue
	}
}

// EventChanSize specifies the size of the event channel used for received synchronously event.
func (_Option) EventChanSize(size int) option.Setting[SubscriberOptions] {
	return func(o *SubscriberOptions) {
		o.EventChanSize = size
	}
}

// EventHandler is the function that will be called to handle the received events.
func (_Option) EventHandler(handler EventHandler) option.Setting[SubscriberOptions] {
	return func(o *SubscriberOptions) {
		o.EventHandler = handler
	}
}

// UnsubscribedCB Unsubscribed callback method.
func (_Option) UnsubscribedCB(cb UnsubscribedCB) option.Setting[SubscriberOptions] {
	return func(o *SubscriberOptions) {
		o.UnsubscribedCB = cb
	}
}
