package broker

import "errors"

// ErrUnsubscribed is an error indicating that the subscriber has been unsubscribed. It is returned by the Subscriber.Next method when the subscriber has been unsubscribed.
var ErrUnsubscribed = errors.New("broker: unsubscribed")

// Subscriber is a convenience return type for the Subscribe method.
type Subscriber interface {
	// Pattern returns the subscription pattern used to create the subscriber.
	Pattern() string
	// QueueName subscribers with the same queue name will create a shared subscription where each receives a subset of messages.
	QueueName() string
	// Unsubscribe unsubscribes the subscriber from the topic.
	Unsubscribe() error
	// Next is a blocking call that waits for the next event to be received from the subscriber.
	Next() (Event, error)
}

// Event is given to a subscription handler for processing
type Event interface {
	// Pattern returns the subscription pattern used to create the event subscriber.
	Pattern() string
	// Topic returns the topic the event was received on.
	Topic() string
	// QueueName subscribers with the same queue name will create a shared subscription where each receives a subset of messages.
	QueueName() string
	// Message returns the raw message payload of the event.
	Message() []byte
	// Ack acknowledges the successful processing of the event. It indicates that the event can be removed from the subscription queue.
	Ack() error
	// Error returns any error that occurred while processing the event, if applicable.
	Error() error
}
