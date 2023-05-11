package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/registry"
	"net"
	"strings"
)

type _RedisWatcher struct {
	ctx       service.Context
	cancel    context.CancelFunc
	keyCache  map[string]string
	eventChan chan *registry.Event
}

func newRedisWatcher(ctx context.Context, r *_RedisRegistry, serviceName string) (watcher registry.Watcher, err error) {
	watchPathList := []string{
		fmt.Sprintf("__keyevent@%d__:set", r.client.Options().DB),
		fmt.Sprintf("__keyevent@%d__:del", r.client.Options().DB),
		fmt.Sprintf("__keyevent@%d__:expired", r.client.Options().DB),
	}

	watch := r.client.Subscribe(ctx)
	if err := watch.Subscribe(ctx, watchPathList...); err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			watch.Close()
		}
	}()

	var keyPath string
	if serviceName != "" {
		keyPath = getServicePath(r.options.KeyPrefix, serviceName)
	} else {
		keyPath = r.options.KeyPrefix + "*"
	}

	keys, err := r.client.Keys(ctx, keyPath).Result()
	if err != nil {
		return nil, err
	}

	keyCache := map[string]string{}

	if len(keys) > 0 {
		values, err := r.client.MGet(ctx, keys...).Result()
		if err != nil {
			return nil, err
		}

		for i, v := range values {
			if v != nil {
				keyCache[keys[i]] = v.(string)
			}
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	eventChan := make(chan *registry.Event, r.options.WatchChanSize)

	go func() {
		<-ctx.Done()
		watch.Close()
	}()

	go func() {
		defer func() {
			close(eventChan)
			for range eventChan {
			}
		}()

		for {
			msg, err := watch.ReceiveMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, redis.ErrClosed) || errors.Is(err, net.ErrClosed) {
					logger.Debugf(r.ctx, "stop watch %q, %s", watchPathList, err)
					return
				}
				logger.Error(r.ctx, "interrupt watch %q, %s", watchPathList, err)
				return
			}

			key := msg.Payload

			if !strings.HasPrefix(key, keyPath[:len(keyPath)-1]) {
				continue
			}

			event := &registry.Event{}

			switch msg.Channel {
			case watchPathList[0]:
				val, err := r.client.Get(ctx, key).Result()
				if err != nil {
					if errors.Is(err, context.Canceled) {
						continue
					}
					logger.Errorf(r.ctx, "get node %q data failed, %s", key, err)
					continue
				}

				_, ok := keyCache[key]
				if ok {
					event.Type = registry.Update
				} else {
					event.Type = registry.Create
				}

				event.Service, err = decodeService([]byte(val))
				if err != nil {
					logger.Errorf(r.ctx, "decode node %q data failed, %s", key, err)
					continue
				}

				keyCache[key] = val

			case watchPathList[1], watchPathList[2]:
				v, ok := keyCache[key]
				if !ok {
					logger.Errorf(r.ctx, "node %q data not cached, %s", key, err)
					continue
				}

				delete(keyCache, key)

				event.Type = registry.Delete
				event.Service, err = decodeService([]byte(v))
				if err != nil {
					logger.Errorf(r.ctx, "decode node %q data failed, %s", key, err)
					continue
				}

			default:
				logger.Errorf(r.ctx, "unknown channel %q", msg.Channel)
				continue
			}

			if len(event.Service.Nodes) <= 0 {
				logger.Debugf(r.ctx, "event service %q node is empty, discard it", event.Service.Name)
				continue
			}

			select {
			case eventChan <- event:
			case <-ctx.Done():
				logger.Debugf(r.ctx, "stop watch %q", watchPathList)
				return
			}
		}
	}()

	return &_RedisWatcher{
		ctx:       r.ctx,
		cancel:    cancel,
		keyCache:  keyCache,
		eventChan: eventChan,
	}, nil
}

// Next is a blocking call
func (w *_RedisWatcher) Next() (*registry.Event, error) {
	for event := range w.eventChan {
		return event, nil
	}
	return nil, registry.ErrWatcherStopped
}

// Stop stop watching
func (w *_RedisWatcher) Stop() {
	w.cancel()
}
