package etcd

import (
	"context"
	"github.com/redis/go-redis/v9"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/registry"
)

type _RedisWatcher struct {
	ctx       service.Context
	stopChan  chan bool
	watchChan <-chan *redis.Message
}

func newRedisWatcher(ctx context.Context, r *_RedisRegistry, serviceName string) (registry.Watcher, error) {
	watchPath := r.options.KeyPrefix
	if serviceName != "" {
		watchPath = getServiceChannel(r.options.KeyPrefix, serviceName)
	}

	stopChan := make(chan bool, 1)
	watch := r.client.PSubscribe(ctx)
	err := watch.PSubscribe(ctx, watchPath)
	if err != nil {
		return nil, err
	}
	watchChan := watch.Channel(redis.WithChannelSize(r.options.WatchChanSize))

	go func() {
		<-stopChan
		watch.Close()
	}()

	return &_RedisWatcher{
		ctx:       r.ctx,
		stopChan:  stopChan,
		watchChan: watchChan,
	}, nil
}

// Next is a blocking call
func (w *_RedisWatcher) Next() (*registry.Event, error) {
	for watchRsp := range w.watchChan {
		redisEvent, err := decodeEvent([]byte(watchRsp.Payload))
		if err != nil {
			logger.Error(w.ctx, err)
			continue
		}

		return &registry.Event{
			Type:    redisEvent.Type,
			Service: redisEvent.Service,
		}, nil
	}

	return nil, registry.ErrWatcherStopped
}

// Stop stop watching
func (w *_RedisWatcher) Stop() {
	select {
	case <-w.stopChan:
		return
	default:
		close(w.stopChan)
	}
}
