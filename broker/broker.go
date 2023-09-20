package broker

import "context"

// Broker is an interface used for asynchronous messaging.
type Broker interface {
	// Publish the data argument to the given topic. The data argument is left untouched and needs to be correctly interpreted on the receiver.
	Publish(ctx context.Context, topic string, data []byte) error
	// Subscribe will express interest in the given topic pattern. Use option EventHandler to handle message events.
	Subscribe(ctx context.Context, pattern string, options ...SubscriberOption) (Subscriber, error)
	// SubscribeSync will express interest in the given topic pattern.
	SubscribeSync(ctx context.Context, pattern string, options ...SubscriberOption) (SyncSubscriber, error)
	// SubscribeChan will express interest in the given topic pattern.
	SubscribeChan(ctx context.Context, pattern string, options ...SubscriberOption) (ChanSubscriber, error)
	// Flush will perform a round trip to the server and return when it receives the internal reply.
	Flush(ctx context.Context) error
	// MaxPayload return max payload bytes.
	MaxPayload() int64
	// Separator return topic path separator.
	Separator() string
}
