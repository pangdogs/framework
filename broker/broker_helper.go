package broker

import (
	"context"
	"kit.golaxy.org/golaxy/service"
)

// Publish the data argument to the given topic. The data argument is left untouched and needs to be correctly interpreted on the receiver.
func Publish(serviceCtx service.Context, ctx context.Context, topic string, data []byte) error {
	return Using(serviceCtx).Publish(ctx, topic, data)
}

// Subscribe will express interest in the given topic pattern. Use option EventHandler to handle message events.
func Subscribe(serviceCtx service.Context, ctx context.Context, pattern string, options ...SubscriberOption) (Subscriber, error) {
	return Using(serviceCtx).Subscribe(ctx, pattern, options...)
}

// SubscribeSync will express interest in the given topic pattern.
func SubscribeSync(serviceCtx service.Context, ctx context.Context, pattern string, options ...SubscriberOption) (SyncSubscriber, error) {
	return Using(serviceCtx).SubscribeSync(ctx, pattern, options...)
}

// SubscribeChan will express interest in the given topic pattern.
func SubscribeChan(serviceCtx service.Context, ctx context.Context, pattern string, options ...SubscriberOption) (ChanSubscriber, error) {
	return Using(serviceCtx).SubscribeChan(ctx, pattern, options...)
}

// Flush will perform a round trip to the server and return when it receives the internal reply.
func Flush(serviceCtx service.Context, ctx context.Context) error {
	return Using(serviceCtx).Flush(ctx)
}

// MaxPayload return max payload bytes.
func MaxPayload(serviceCtx service.Context) int64 {
	return Using(serviceCtx).MaxPayload()
}

// Separator return topic path separator.
func Separator(serviceCtx service.Context) string {
	return Using(serviceCtx).Separator()
}
