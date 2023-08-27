package etcd_registry

import (
	"context"
	etcd_client "go.etcd.io/etcd/client/v3"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/registry"
)

type _EtcdWatcher struct {
	ctx       service.Context
	cancel    context.CancelFunc
	watchChan etcd_client.WatchChan
	eventChan chan *registry.Event
	client    *etcd_client.Client
}

func newEtcdWatcher(ctx context.Context, r *_EtcdRegistry, serviceName string) (registry.Watcher, error) {
	watchPath := r.options.KeyPrefix
	if serviceName != "" {
		watchPath = getServicePath(r.options.KeyPrefix, serviceName)
	}

	ctx, cancel := context.WithCancel(ctx)
	watchChan := r.client.Watch(ctx, watchPath, etcd_client.WithPrefix(), etcd_client.WithPrevKV())
	eventChan := make(chan *registry.Event, r.options.WatchChanSize)

	go func() {
		defer func() {
			close(eventChan)
			for range eventChan {
			}
		}()

		logger.Debugf(r.ctx, "start watch %q", watchPath)

		for watchRsp := range watchChan {
			if watchRsp.Canceled {
				logger.Debugf(r.ctx, "stop watch %q", watchPath)
				return
			}
			if watchRsp.Err() != nil {
				logger.Error(r.ctx, "interrupt watch %q, %s", watchPath, watchRsp.Err())
				return
			}

			for _, etcdEvent := range watchRsp.Events {
				event := &registry.Event{}
				var err error

				switch etcdEvent.Type {
				case etcd_client.EventTypePut:
					if etcdEvent.IsCreate() {
						event.Type = registry.Create
					} else if etcdEvent.IsModify() {
						event.Type = registry.Update
					}

					// get service from Kv
					event.Service, err = decodeService(etcdEvent.Kv.Value)

				case etcd_client.EventTypeDelete:
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
		cancel:    cancel,
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
	return nil, registry.ErrStoppedWatching
}

// Stop stop watching
func (w *_EtcdWatcher) Stop() {
	w.cancel()
}
