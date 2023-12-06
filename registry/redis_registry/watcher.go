package redis_registry

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"kit.golaxy.org/plugins/log"
	"kit.golaxy.org/plugins/registry"
	"net"
	"strings"
)

func (r *_Registry) newWatcher(ctx context.Context, pattern string) (watcher *_Watcher, err error) {
	var keyPath string
	if pattern != "" {
		keyPath = getServicePath(r.options.KeyPrefix, pattern)
	} else {
		keyPath = r.options.KeyPrefix + "*"
	}

	watchPathPrefixList := []string{
		fmt.Sprintf("__keyspace@%d__:", r.client.Options().DB),
		fmt.Sprintf("__keyevent@%d__:", r.client.Options().DB),
	}

	watchPathList := []string{
		watchPathPrefixList[0] + keyPath,
		watchPathPrefixList[1] + "expired",
	}

	watch := r.client.PSubscribe(ctx)
	defer func() {
		if err != nil {
			watch.Close()
		}
	}()

	if err := watch.PSubscribe(ctx, watchPathList...); err != nil {
		return nil, fmt.Errorf("%w: %w", registry.ErrRegistry, err)
	}

	keys, err := r.client.Keys(ctx, keyPath).Result()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", registry.ErrRegistry, err)
	}

	keyCache := map[string]string{}

	if len(keys) > 0 {
		values, err := r.client.MGet(ctx, keys...).Result()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", registry.ErrRegistry, err)
		}

		for i, v := range values {
			if v != nil {
				keyCache[keys[i]] = v.(string)
			}
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	eventChan := make(chan *registry.Event, r.options.WatchChanSize)

	watcher = &_Watcher{
		registry:       r,
		ctx:            ctx,
		cancel:         cancel,
		stoppedChan:    make(chan struct{}, 1),
		pattern:        keyPath,
		pathPrefixList: watchPathPrefixList,
		pathList:       watchPathList,
		keyCache:       keyCache,
		pubSub:         watch,
		eventChan:      eventChan,
	}

	go watcher.mainLoop()

	return watcher, nil
}

type _Watcher struct {
	registry       *_Registry
	ctx            context.Context
	cancel         context.CancelFunc
	stoppedChan    chan struct{}
	pattern        string
	pathPrefixList []string
	pathList       []string
	keyCache       map[string]string
	pubSub         *redis.PubSub
	eventChan      chan *registry.Event
}

// Pattern watching pattern
func (w *_Watcher) Pattern() string {
	return strings.TrimPrefix(w.pattern, w.registry.options.KeyPrefix)
}

// Next is a blocking call
func (w *_Watcher) Next() (*registry.Event, error) {
	for event := range w.eventChan {
		return event, nil
	}
	return nil, registry.ErrStoppedWatching
}

// Stop stop watching
func (w *_Watcher) Stop() <-chan struct{} {
	w.cancel()
	return w.stoppedChan
}

func (w *_Watcher) mainLoop() {
	defer func() {
		close(w.eventChan)
		w.stoppedChan <- struct{}{}
	}()

	log.Debugf(w.registry.ctx, "start watch %q", w.pathList)

loop:
	for {
		msg, err := w.pubSub.ReceiveMessage(w.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, redis.ErrClosed) || errors.Is(err, net.ErrClosed) {
				log.Debugf(w.registry.ctx, "stop watch %q, %s", w.pathList, err)
				break loop
			}
			log.Errorf(w.registry.ctx, "interrupt watch %q, %s", w.pathList, err)
			break loop
		}

		var key, opt string

		switch msg.Pattern {
		case w.pathList[0]:
			key = strings.TrimPrefix(msg.Channel, w.pathPrefixList[0])
			opt = msg.Payload
		case w.pathList[1]:
			key = msg.Payload
			opt = strings.TrimPrefix(msg.Channel, w.pathPrefixList[1])
		default:
			continue
		}

		if !strings.HasPrefix(key, w.pattern[:len(w.pattern)-1]) {
			continue
		}

		event := &registry.Event{}

		switch opt {
		case "set":
			val, err := w.registry.client.Get(w.ctx, key).Result()
			if err != nil {
				if errors.Is(err, context.Canceled) {
					continue
				}
				log.Errorf(w.registry.ctx, "get service node %q data failed, %s", key, err)
				continue
			}

			_, ok := w.keyCache[key]
			if ok {
				event.Type = registry.Update
			} else {
				event.Type = registry.Create
			}

			event.Service, err = decodeService([]byte(val))
			if err != nil {
				log.Errorf(w.registry.ctx, "decode service %q data failed, %s", val, err)
				continue
			}

			w.keyCache[key] = val

		case "del", "expired":
			v, ok := w.keyCache[key]
			if !ok {
				log.Errorf(w.registry.ctx, "service node %q data not cached", key)
				continue
			}

			delete(w.keyCache, key)

			event.Type = registry.Delete
			event.Service, err = decodeService([]byte(v))
			if err != nil {
				log.Errorf(w.registry.ctx, "decode service %q data failed, %s", v, err)
				continue
			}

		default:
			continue
		}

		if len(event.Service.Nodes) <= 0 {
			log.Warnf(w.registry.ctx, "event service %q node is empty, discard it", event.Service.Name)
			continue
		}

		select {
		case w.eventChan <- event:
		case <-w.ctx.Done():
			log.Debugf(w.registry.ctx, "stop watch %q", w.pathList)
			break loop
		}
	}

	if err := w.pubSub.Close(); err != nil {
		log.Errorf(w.registry.ctx, "close watch %q failed, %s", w.pathList, err)
	}
}
