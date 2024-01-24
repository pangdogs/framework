package broker

import (
	"context"
)

// ISubscriber is a convenience return type for the IBroker.Subscribe method.
type ISubscriber interface {
	context.Context
	// Pattern returns the subscription pattern used to create the subscriber.
	Pattern() string
	// Queue subscribers with the same queue name will create a shared subscription where each receives a subset of messages.
	Queue() string
	// Unsubscribe unsubscribes the subscriber from the topic.
	Unsubscribe() <-chan struct{}
}

// ISyncSubscriber is a convenience return type for the IBroker.SubscribeSync method.
type ISyncSubscriber interface {
	ISubscriber
	// Next is a blocking call that waits for the next event to be received from the subscriber.
	Next() (IEvent, error)
}

// IChanSubscriber is a convenience return type for the IBroker.SubscribeChan method.
type IChanSubscriber interface {
	ISubscriber
	// EventChan returns a channel that can be used to receive events from the subscriber.
	EventChan() (<-chan IEvent, error)
}

// IEvent is given to a subscription handler for processing.
type IEvent interface {
	// Pattern returns the subscription pattern used to create the event subscriber.
	Pattern() string
	// Topic returns the topic the event was received on.
	Topic() string
	// Queue subscribers with the same queue name will create a shared subscription where each receives a subset of messages.
	Queue() string
	// Message returns the raw message payload of the event.
	Message() []byte
	// Ack acknowledges the successful processing of the event. It indicates that the event can be removed from the subscription queue.
	Ack(ctx context.Context) error
	// Nak negatively acknowledges a message. This tells the server to redeliver the message.
	Nak(ctx context.Context) error
}
