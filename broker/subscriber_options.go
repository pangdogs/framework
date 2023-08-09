package broker

// Option is a helper struct to provide default options.
type Option struct{}

type (
	// EventHandler is used to process messages via a subscription of a topic. The handler is passed a publication interface which contains the
	// message and optional Ack method to acknowledge receipt of the message.
	EventHandler = func(e Event) error
	// UnsubscribedCb Unsubscribed callback method.
	UnsubscribedCb = func(sub Subscriber)
)

// SubscriberOptions represents the options for subscribe topic.
type SubscriberOptions struct {
	// AutoAck defaults to true. When a handler returns with a nil error the message is acked.
	AutoAck bool
	// QueueName subscribers with the same queue name will create a shared subscription where each receives a subset of messages.
	QueueName string
	// EventHandler is the function that will be called to handle the received events.
	EventHandler EventHandler
	// EventChanSize specifies the size of the event channel used for received synchronously event.
	EventChanSize int
	// UnsubscribedCb Unsubscribed callback method.
	UnsubscribedCb UnsubscribedCb
}

// SubscriberOption represents a configuration option for subscribe topic.
type SubscriberOption func(*SubscriberOptions)

// Default sets the default options for subscribe topic.
func (Option) Default() SubscriberOption {
	return func(options *SubscriberOptions) {
		Option{}.AutoAck(true)(options)
		Option{}.QueueName("")(options)
		Option{}.EventHandler(nil)(options)
		Option{}.EventChanSize(128)(options)
		Option{}.UnsubscribedCb(nil)(options)
	}
}

// AutoAck defaults to true. When a handler returns with a nil error the message is acked.
func (Option) AutoAck(b bool) SubscriberOption {
	return func(o *SubscriberOptions) {
		o.AutoAck = b
	}
}

// QueueName subscribers with the same queue name will create a shared subscription where each
// receives a subset of messages.
func (Option) QueueName(name string) SubscriberOption {
	return func(o *SubscriberOptions) {
		o.QueueName = name
	}
}

// EventHandler is the function that will be called to handle the received events.
func (Option) EventHandler(handler EventHandler) SubscriberOption {
	return func(o *SubscriberOptions) {
		o.EventHandler = handler
	}
}

// EventChanSize specifies the size of the event channel used for received synchronously event.
func (Option) EventChanSize(size int) SubscriberOption {
	return func(o *SubscriberOptions) {
		o.EventChanSize = size
	}
}

// UnsubscribedCb Unsubscribed callback method.
func (Option) UnsubscribedCb(fn UnsubscribedCb) SubscriberOption {
	return func(o *SubscriberOptions) {
		o.UnsubscribedCb = fn
	}
}
