package etcd

import (
	"context"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/registry"
)

type _EtcdWatcher struct {
	ctx       service.Context
	stopChan  chan bool
	watchChan clientv3.WatchChan
	eventChan chan *registry.Event
	client    *clientv3.Client
}

func newEtcdWatcher(ctx context.Context, r *_EtcdRegistry, serviceName string) (registry.Watcher, error) {
	watchPath := r.options.KeyPrefix
	if serviceName != "" {
		watchPath = getServicePath(r.options.KeyPrefix, serviceName)
	}

	ctx, cancel := context.WithCancel(ctx)
	stopChan := make(chan bool, 1)
	watchChan := r.client.Watch(ctx, watchPath, clientv3.WithPrefix(), clientv3.WithPrevKV())
	eventChan := make(chan *registry.Event, r.options.WatchChanSize)

	go func() {
		<-stopChan
		cancel()
		for range eventChan {
		}
	}()

	go func() {
		defer func() {
			close(eventChan)
			select {
			case <-stopChan:
			default:
				close(stopChan)
			}
		}()

		for watchRsp := range watchChan {
			if watchRsp.Err() != nil {
				if errors.Is(watchRsp.Err(), context.Canceled) {
					logger.Debugf(r.ctx, "stop watch %q", watchPath)
					return
				}
				logger.Error(r.ctx, watchRsp.Err())
				return
			}
			if watchRsp.Canceled {
				logger.Debugf(r.ctx, "stop watch %q", watchPath)
				return
			}

			for _, etcdEvent := range watchRsp.Events {
				event := &registry.Event{}
				var err error

				switch etcdEvent.Type {
				case clientv3.EventTypePut:
					if etcdEvent.IsCreate() {
						event.Type = registry.Create
					} else if etcdEvent.IsModify() {
						event.Type = registry.Update
					}

					// get service from Kv
					event.Service, err = decodeService(etcdEvent.Kv.Value)

				case clientv3.EventTypeDelete:
					event.Type = registry.Delete

					// get service from prevKv
					event.Service, err = decodeService(etcdEvent.PrevKv.Value)

				default:
					logger.Errorf(r.ctx, "unknown event type %q", etcdEvent.Type)
					continue
				}

				if err != nil {
					logger.Error(r.ctx, err)
					continue
				}

				if len(event.Service.Nodes) <= 0 {
					logger.Debugf(r.ctx, "event service %q node is empty, discard it", event.Service.Name)
					continue
				}

				select {
				case eventChan <- event:
				case <-ctx.Done():
					logger.Debugf(r.ctx, "stop watch %q", watchPath)
					return
				}
			}
		}
	}()

	return &_EtcdWatcher{
		ctx:       r.ctx,
		stopChan:  stopChan,
		watchChan: watchChan,
		eventChan: eventChan,
		client:    r.client,
	}, nil
}

// Next is a blocking call
func (w *_EtcdWatcher) Next() (*registry.Event, error) {
	for event := range w.eventChan {
		return event, nil
	}
	return nil, registry.ErrWatcherStopped
}

// Stop stop watching
func (w *_EtcdWatcher) Stop() {
	select {
	case <-w.stopChan:
	default:
		close(w.stopChan)
	}
}
