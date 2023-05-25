package broker

import (
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy/service"
)

// Publish the data argument to the given topic. The data argument is left untouched and needs to be correctly interpreted on the receiver.
func Publish(serviceCtx service.Context, ctx context.Context, topic string, data []byte) error {
	return Get(serviceCtx).Publish(ctx, topic, data)
}

// Subscribe will express interest in the given topic pattern. Use option EventHandler to handle message events.
func Subscribe(serviceCtx service.Context, ctx context.Context, pattern string, options ...SubscriberOption) (Subscriber, error) {
	return Get(serviceCtx).Subscribe(ctx, pattern, options...)
}

// SubscribeSync will express interest in the given topic pattern.
func SubscribeSync(serviceCtx service.Context, ctx context.Context, pattern string, options ...SubscriberOption) (SyncSubscriber, error) {
	return Get(serviceCtx).SubscribeSync(ctx, pattern, options...)
}

// SubscribeChan will express interest in the given topic pattern.
func SubscribeChan(serviceCtx service.Context, ctx context.Context, pattern string, options ...SubscriberOption) (ChanSubscriber, error) {
	return Get(serviceCtx).SubscribeChan(ctx, pattern, options...)
}

// Flush will perform a round trip to the server and return when it receives the internal reply.
func Flush(serviceCtx service.Context, ctx context.Context) error {
	return Get(serviceCtx).Flush(ctx)
}

// MaxPayload return max payload bytes.
func MaxPayload(serviceCtx service.Context) int64 {
	return Get(serviceCtx).MaxPayload()
}
