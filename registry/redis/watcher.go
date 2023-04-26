package etcd

import (
	"context"
	"github.com/redis/go-redis/v9"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/registry"
	"time"
)

type _RedisWatcher struct {
	ctx       service.Context
	watch     *redis.PubSub
	watchChan <-chan *redis.Message
	timeout   time.Duration
}

func newRedisWatcher(ctx context.Context, r *_RedisRegistry, timeout time.Duration, serviceName string) (registry.Watcher, error) {
	ctx, cancel := context.WithCancel(ctx)
	stop := make(chan bool, 1)

	go func() {
		<-stop
		cancel()
	}()

	watchPath := r.options.KeyPrefix
	if serviceName != "" {
		watchPath = getServicePath(r.options.KeyPrefix, serviceName)
	}

	watch := r.client.PSubscribe(ctx, watchPath)

	return &_RedisWatcher{
		ctx:       r.ctx,
		watch:     watch,
		watchChan: watch.Channel(),
		timeout:   timeout,
	}, nil
}

// Next is a blocking call
func (ew *_RedisWatcher) Next() (*registry.Result, error) {
	for msg := range ew.watchChan {

	}

	return nil, registry.ErrWatcherStopped
}

// Stop stop watching
func (ew *_RedisWatcher) Stop() {
	ew.watch.Close()
}
