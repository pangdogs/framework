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

package discovery_etcd

import (
	"context"
	"errors"
	"fmt"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/addins/discovery"
	"git.golaxy.org/framework/addins/log"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// addWatcher 创建 ETCD 前缀 watcher，并将变化投递到事件流或回调。
func (r *_EtcdRegistry) addWatcher(ctx context.Context, pattern string, handler discovery.EventHandler, revision int64) (<-chan discovery.Event, async.Signal, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-r.scope.Context().Done():
		return nil, async.Signal{}, errors.New("registry: registry is terminating")
	default:
	}

	if !r.barrier.Join(1) {
		return nil, async.Signal{}, errors.New("registry: registry is terminating")
	}
	defer r.barrier.Done()

	key := r.options.KeyPrefix
	if pattern != "" {
		key = r.newServiceKey(pattern)
	}

	var eventChan *generic.UnboundedChannel[discovery.Event]
	if handler != nil {
		eventChan = generic.NewUnboundedChannel[discovery.Event]()
	}

	handleEvent := func(event discovery.Event) {
		if eventChan != nil {
			eventChan.In() <- event
		}
		if handler != nil {
			handler.Call(r.svcCtx.AutoRecover(), r.svcCtx.ReportError(), func(panicErr error) bool {
				if panicErr != nil {
					log.L(r.svcCtx).Error("handle event from watching etcd key panicked",
						zap.String("key", key),
						zap.Int64("revision", revision),
						zap.Error(panicErr))
				}
				return false
			}, event)
		}
	}

	stopped, stoppedSignal := async.NewSignal()
	finish := func() {
		if eventChan != nil {
			eventChan.Close()
		}
		stopped.Complete()
	}

	future := async.SpawnVoid(r.scope, func(scopeCtx context.Context) {
		ctx, cancel := context.WithCancel(ctx)
		stopOwner := context.AfterFunc(scopeCtx, cancel)
		defer finish()
		defer cancel()
		defer stopOwner()

		log.L(r.svcCtx).Debug("watching for service changes started", zap.String("key", key), zap.Int64("revision", revision))

		for watchRsp := range r.client.Watch(ctx, key, etcdv3.WithPrefix(), etcdv3.WithPrevKV(), etcdv3.WithRev(revision)) {
			if watchRsp.Canceled {
				log.L(r.svcCtx).Debug("watching etcd key canceled", zap.String("key", key), zap.Int64("revision", revision), zap.Error(watchRsp.Err()))
				break
			}
			if watchRsp.Err() != nil {
				log.L(r.svcCtx).Error("watching etcd key unexpectedly interrupted", zap.String("key", key), zap.Int64("revision", revision), zap.Error(watchRsp.Err()))
				handleEvent(discovery.Event{Type: discovery.EventType_Error, Error: fmt.Errorf("registry: %w", watchRsp.Err())})
				break
			}

			for _, etcdEvent := range watchRsp.Events {
				var event discovery.Event
				var err error

				switch etcdEvent.Type {
				case etcdv3.EventTypePut:
					if etcdEvent.IsCreate() {
						event.Type = discovery.EventType_Create
					} else if etcdEvent.IsModify() {
						event.Type = discovery.EventType_Update
					}

					event.Service, err = decodeService(etcdEvent.Kv.Value)
					if err != nil {
						log.L(r.svcCtx).Error("decode service failed", zap.ByteString("key", etcdEvent.Kv.Key), zap.Error(err))
						continue
					}

				case etcdv3.EventTypeDelete:
					event.Type = discovery.EventType_Delete

					event.Service, err = decodeService(etcdEvent.PrevKv.Value)
					if err != nil {
						log.L(r.svcCtx).Error("decode service failed", zap.ByteString("key", etcdEvent.PrevKv.Key), zap.Error(err))
						continue
					}

				default:
					log.L(r.svcCtx).Error("unknown event type", zap.String("type", etcdEvent.Type.String()))
					continue
				}

				if len(event.Service.Nodes) <= 0 {
					log.L(r.svcCtx).Warn("event service node is empty, discard it", zap.String("service", event.Service.Name))
					continue
				}

				event.Service.Revision = watchRsp.Header.Revision

				handleEvent(event)
			}
		}

		log.L(r.svcCtx).Debug("watching for service changes stopped", zap.String("key", key), zap.Int64("revision", revision))
	})
	future.OnComplete(func(ret async.Result) {
		if ret.Error != nil && !errors.Is(ret.Error, async.ErrScopeClosed) {
			log.L(r.svcCtx).Error("registry watcher task failed", zap.Error(ret.Error))
		}
	})

	if ret, ok := future.TryGet(); ok && errors.Is(ret.Error, async.ErrScopeClosed) {
		finish()
		return nil, async.Signal{}, errors.New("registry: registry is terminating")
	}

	if eventChan != nil {
		return eventChan.Out(), stoppedSignal, nil
	}
	return nil, stoppedSignal, nil
}
