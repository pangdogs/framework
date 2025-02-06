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

package etcd_discovery

import (
	"context"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/addins/discovery"
	"git.golaxy.org/framework/addins/log"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"strings"
)

func (r *_Registry) newWatcher(ctx context.Context, pattern string, revision ...int64) (*_Watcher, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	watchKey := r.options.KeyPrefix
	if pattern != "" {
		watchKey = getServicePath(r.options.KeyPrefix, pattern)
	}

	watchOpts := []etcdv3.OpOption{etcdv3.WithPrefix(), etcdv3.WithPrevKV()}
	if len(revision) > 0 {
		watchOpts = append(watchOpts, etcdv3.WithRev(revision[0]))
	}

	watcher := &_Watcher{
		registry:   r,
		ctx:        ctx,
		terminate:  cancel,
		terminated: async.MakeAsyncRet(),
		pattern:    pattern,
		watchChan:  r.client.Watch(ctx, watchKey, watchOpts...),
		eventChan:  make(chan *discovery.Event, r.options.WatchChanSize),
	}

	go watcher.mainLoop()

	return watcher, nil
}

type _Watcher struct {
	registry   *_Registry
	ctx        context.Context
	terminate  context.CancelFunc
	terminated chan async.Ret
	pattern    string
	watchChan  etcdv3.WatchChan
	eventChan  chan *discovery.Event
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

	log.Debugf(w.registry.svcCtx, "start watch %q", w.pattern)

	for watchRsp := range w.watchChan {
		if watchRsp.Canceled {
			log.Debugf(w.registry.svcCtx, "stop watch %q", w.pattern)
			return
		}
		if watchRsp.Err() != nil {
			log.Errorf(w.registry.svcCtx, "interrupt watch %q, %s", w.pattern, watchRsp.Err())
			return
		}

		for _, event := range watchRsp.Events {
			ret := &discovery.Event{}
			var err error

			switch event.Type {
			case etcdv3.EventTypePut:
				if event.IsCreate() {
					ret.Type = discovery.Create
				} else if event.IsModify() {
					ret.Type = discovery.Update
				}

				// get service from Kv
				ret.Service, err = decodeService(event.Kv.Value)
				if err != nil {
					log.Errorf(w.registry.svcCtx, "decode service %q failed, %s", event.Kv.Value, err)
					continue
				}

			case etcdv3.EventTypeDelete:
				ret.Type = discovery.Delete

				// get service from prevKv
				ret.Service, err = decodeService(event.PrevKv.Value)
				if err != nil {
					log.Errorf(w.registry.svcCtx, "decode service %q failed, %s", event.PrevKv.Value, err)
					continue
				}

			default:
				log.Errorf(w.registry.svcCtx, "unknown event type %q", event.Type)
				continue
			}

			if len(ret.Service.Nodes) <= 0 {
				log.Warnf(w.registry.svcCtx, "event service %q node is empty, discard it", ret.Service.Name)
				continue
			}

			ret.Service.Revision = watchRsp.Header.Revision

			select {
			case w.eventChan <- ret:
			case <-w.ctx.Done():
				w.terminate()
			}
		}
	}
}
