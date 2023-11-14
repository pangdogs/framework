package broker

import (
	"context"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/option"
)

// Publish the data argument to the given topic. The data argument is left untouched and needs to be correctly interpreted on the receiver.
func Publish(servCtx service.Context, ctx context.Context, topic string, data []byte) error {
	return Using(servCtx).Publish(ctx, topic, data)
}

// Subscribe will express interest in the given topic pattern. Use option EventHandler to handle message events.
func Subscribe(servCtx service.Context, ctx context.Context, pattern string, settings ...option.Setting[SubscriberOptions]) (Subscriber, error) {
	return Using(servCtx).Subscribe(ctx, pattern, settings...)
}

// SubscribeSync will express interest in the given topic pattern.
func SubscribeSync(servCtx service.Context, ctx context.Context, pattern string, settings ...option.Setting[SubscriberOptions]) (SyncSubscriber, error) {
	return Using(servCtx).SubscribeSync(ctx, pattern, settings...)
}

// SubscribeChan will express interest in the given topic pattern.
func SubscribeChan(servCtx service.Context, ctx context.Context, pattern string, settings ...option.Setting[SubscriberOptions]) (ChanSubscriber, error) {
	return Using(servCtx).SubscribeChan(ctx, pattern, settings...)
}

// Flush will perform a round trip to the server and return when it receives the internal reply.
func Flush(servCtx service.Context, ctx context.Context) error {
	return Using(servCtx).Flush(ctx)
}

// MaxPayload return max payload bytes.
func MaxPayload(servCtx service.Context) int64 {
	return Using(servCtx).MaxPayload()
}

// Separator return topic path separator.
func Separator(servCtx service.Context) string {
	return Using(servCtx).Separator()
}
