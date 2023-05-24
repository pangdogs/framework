package broker

// WithOption is a helper struct to provide default options.
type WithOption struct{}

// EventHandler is used to process messages via a subscription of a topic. The handler is passed a publication interface which contains the
// message and optional Ack method to acknowledge receipt of the message.
type EventHandler = func(e Event) error

// SubscriberOptions represents the options for subscribe topic.
type SubscriberOptions struct {
	// AutoAck defaults to true. When a handler returns with a nil error the message is acked.
	AutoAck bool
	// QueueName subscribers with the same queue name will create a shared subscription where each
	// receives a subset of messages.
	QueueName string
	// EventHandler is the function that will be called to handle the received events. If EventHandler is set to nil, messages will
	// be received synchronously using Subscription.Next().
	EventHandler EventHandler
	// EventChanSize specifies the size of the event channel used for received synchronously event.
	EventChanSize int
}

// SubscriberOption represents a configuration option for subscribe topic.
type SubscriberOption func(*SubscriberOptions)

// Default sets the default options for subscribe topic.
func (WithOption) Default() SubscriberOption {
	return func(options *SubscriberOptions) {
		WithOption{}.AutoAck(true)(options)
		WithOption{}.QueueName("")(options)
		WithOption{}.EventHandler(nil)(options)
		WithOption{}.EventChanSize(128)(options)
	}
}

// AutoAck defaults to true. When a handler returns with a nil error the message is acked.
func (WithOption) AutoAck(b bool) SubscriberOption {
	return func(o *SubscriberOptions) {
		o.AutoAck = b
	}
}

// QueueName subscribers with the same queue name will create a shared subscription where each
// receives a subset of messages.
func (WithOption) QueueName(name string) SubscriberOption {
	return func(o *SubscriberOptions) {
		o.QueueName = name
	}
}

// EventHandler is the function that will be called to handle the received events. If EventHandler is set to nil, messages will
// be received synchronously using Subscription.Next().
func (WithOption) EventHandler(handler EventHandler) SubscriberOption {
	return func(o *SubscriberOptions) {
		o.EventHandler = handler
	}
}

// EventChanSize specifies the size of the event channel used for received synchronously event.
func (WithOption) EventChanSize(size int) SubscriberOption {
	return func(o *SubscriberOptions) {
		o.EventChanSize = size
	}
}
