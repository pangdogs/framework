package broker

import (
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
)

type (
	// EventHandler is used to process messages via a subscription of a topic. The handler is passed a publication interface which contains the
	// message and optional Ack method to acknowledge receipt of the message.
	EventHandler = generic.DelegateFunc1[IEvent, error]
	// UnsubscribedCB Unsubscribed callback method.
	UnsubscribedCB = generic.DelegateAction1[ISubscriber]
)

// SubscriberOptions represents the options for subscribe topic.
type SubscriberOptions struct {
	// AutoAck defaults to true. When a handler returns with a nil error the message is acked.
	AutoAck bool
	// Queue subscribers with the same queue name will create a shared subscription where each receives a subset of messages.
	Queue string
	// EventHandler is the function that will be called to handle the received events.
	EventHandler EventHandler
	// EventChanSize specifies the size of the event channel used for received synchronously event.
	EventChanSize int
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
		With.EventHandler(nil)(options)
		With.EventChanSize(128)(options)
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

// EventHandler is the function that will be called to handle the received events.
func (_Option) EventHandler(handler EventHandler) option.Setting[SubscriberOptions] {
	return func(o *SubscriberOptions) {
		o.EventHandler = handler
	}
}

// EventChanSize specifies the size of the event channel used for received synchronously event.
func (_Option) EventChanSize(size int) option.Setting[SubscriberOptions] {
	return func(o *SubscriberOptions) {
		o.EventChanSize = size
	}
}

// UnsubscribedCB Unsubscribed callback method.
func (_Option) UnsubscribedCB(handler UnsubscribedCB) option.Setting[SubscriberOptions] {
	return func(o *SubscriberOptions) {
		o.UnsubscribedCB = handler
	}
}
