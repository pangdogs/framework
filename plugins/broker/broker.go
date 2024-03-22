package broker

import (
	"context"
	"errors"
	"git.golaxy.org/core/util/option"
)

var (
	// ErrUnsubscribed is an error indicating that the subscriber has been unsubscribed. It is returned by the ISyncSubscriber.Next method when the subscriber has been unsubscribed.
	ErrUnsubscribed = errors.New("broker: unsubscribed")
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
	// SubscribeSync will express interest in the given topic pattern.
	SubscribeSync(ctx context.Context, pattern string, settings ...option.Setting[SubscriberOptions]) (ISyncSubscriber, error)
	// SubscribeChan will express interest in the given topic pattern.
	SubscribeChan(ctx context.Context, pattern string, settings ...option.Setting[SubscriberOptions]) (IChanSubscriber, error)
	// Flush will perform a round trip to the server and return when it receives the internal reply.
	Flush(ctx context.Context) error
	// GetDeliveryReliability return message delivery reliability.
	GetDeliveryReliability() DeliveryReliability
	// GetMaxPayload return max payload bytes.
	GetMaxPayload() int64
	// GetSeparator return topic path separator.
	GetSeparator() string
}
