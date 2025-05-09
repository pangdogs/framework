/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package redis_discovery

import (
	"context"
	"errors"
	"fmt"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/addins/discovery"
	"git.golaxy.org/framework/addins/log"
	"github.com/redis/go-redis/v9"
	"net"
	"strings"
)

func (r *_Registry) newWatcher(ctx context.Context, pattern string) (watcher *_Watcher, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

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

	pubSub := r.client.PSubscribe(ctx)
	defer func() {
		if watcher == nil {
			pubSub.Close()
		}
	}()

	if err := pubSub.PSubscribe(ctx, watchPathList...); err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}

	keys, err := r.client.Keys(ctx, keyPath).Result()
	if err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}

	keyCache := map[string]string{}

	if len(keys) > 0 {
		values, err := r.client.MGet(ctx, keys...).Result()
		if err != nil {
			return nil, fmt.Errorf("registry: %w", err)
		}

		for i, v := range values {
			if v != nil {
				keyCache[keys[i]] = v.(string)
			}
		}
	}

	eventChan := make(chan *discovery.Event, r.options.WatchChanSize)

	watcher = &_Watcher{
		Context:        ctx,
		registry:       r,
		terminate:      cancel,
		terminated:     async.MakeAsyncRet(),
		pattern:        keyPath,
		pathPrefixList: watchPathPrefixList,
		pathList:       watchPathList,
		keyCache:       keyCache,
		pubSub:         pubSub,
		eventChan:      eventChan,
	}

	go watcher.mainLoop()

	return watcher, nil
}

type _Watcher struct {
	context.Context
	registry       *_Registry
	terminate      context.CancelFunc
	terminated     chan async.Ret
	pattern        string
	pathPrefixList []string
	pathList       []string
	keyCache       map[string]string
	pubSub         *redis.PubSub
	eventChan      chan *discovery.Event
}

// Pattern watching pattern
func (w *_Watcher) Pattern() string {
	return strings.TrimPrefix(w.pattern, w.registry.options.KeyPrefix)
}

// Next is a blocking call
func (w *_Watcher) Next() (*discovery.Event, error) {
	for event := range w.eventChan {
		return event, nil
	}
	return nil, discovery.ErrTerminated
}

// Terminate stop watching
func (w *_Watcher) Terminate() async.AsyncRet {
	w.terminate()
	return w.terminated
}

// Terminated stopped notify
func (w *_Watcher) Terminated() async.AsyncRet {
	return w.terminated
}

func (w *_Watcher) mainLoop() {
	defer func() {
		w.terminate()
		close(w.eventChan)
		async.Return(w.terminated, async.VoidRet)
	}()

	log.Debugf(w.registry.svcCtx, "start watch %q", w.pathList)

	go func() {
		<-w.Done()
		w.pubSub.Close()
	}()

loop:
	for {
		msg, err := w.pubSub.ReceiveMessage(context.Background())
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Debugf(w.registry.svcCtx, "stop watch %q, %s", w.pathList, err)
				break loop
			}
			log.Errorf(w.registry.svcCtx, "interrupt watch %q, %s", w.pathList, err)
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

		event := &discovery.Event{}

		switch opt {
		case "set":
			val, err := w.registry.client.Get(w, key).Result()
			if err != nil {
				if errors.Is(err, context.Canceled) {
					continue
				}
				log.Errorf(w.registry.svcCtx, "get service node %q data failed, %s", key, err)
				continue
			}

			_, ok := w.keyCache[key]
			if ok {
				event.Type = discovery.Update
			} else {
				event.Type = discovery.Create
			}

			event.Service, err = decodeService([]byte(val))
			if err != nil {
				log.Errorf(w.registry.svcCtx, "decode service %q data failed, %s", val, err)
				continue
			}

			w.keyCache[key] = val

		case "del", "expired":
			v, ok := w.keyCache[key]
			if !ok {
				log.Errorf(w.registry.svcCtx, "service node %q data not cached", key)
				continue
			}

			delete(w.keyCache, key)

			event.Type = discovery.Delete
			event.Service, err = decodeService([]byte(v))
			if err != nil {
				log.Errorf(w.registry.svcCtx, "decode service %q data failed, %s", v, err)
				continue
			}

		default:
			continue
		}

		if len(event.Service.Nodes) <= 0 {
			log.Warnf(w.registry.svcCtx, "event service %q node is empty, discard it", event.Service.Name)
			continue
		}

		select {
		case w.eventChan <- event:
		case <-w.Done():
			log.Debugf(w.registry.svcCtx, "stop watch %q", w.pathList)
			break loop
		}
	}

	w.pubSub.Close()
}
