package etcd_registry

import (
	"context"
	etcd_client "go.etcd.io/etcd/client/v3"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/log"
	"kit.golaxy.org/plugins/registry"
)

func (r *_Registry) newWatcher(ctx context.Context, serviceName string) (registry.Watcher, error) {
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

		log.Debugf(r.ctx, "start watch %q", watchPath)

		for watchRsp := range watchChan {
			if watchRsp.Canceled {
				log.Debugf(r.ctx, "stop watch %q", watchPath)
				return
			}
			if watchRsp.Err() != nil {
				log.Errorf(r.ctx, "interrupt watch %q, %s", watchPath, watchRsp.Err())
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
					log.Errorf(r.ctx, "unknown event type %q", etcdEvent.Type)
					continue
				}

				if err != nil {
					log.Error(r.ctx, err)
					continue
				}

				if len(event.Service.Nodes) <= 0 {
					log.Debugf(r.ctx, "event service %q node is empty, discard it", event.Service.Name)
					continue
				}

				select {
				case eventChan <- event:
				case <-ctx.Done():
					log.Debugf(r.ctx, "stop watch %q", watchPath)
					return
				}
			}
		}
	}()

	return &_Watcher{
		ctx:       r.ctx,
		cancel:    cancel,
		watchChan: watchChan,
		eventChan: eventChan,
		client:    r.client,
	}, nil
}

type _Watcher struct {
	ctx       service.Context
	cancel    context.CancelFunc
	watchChan etcd_client.WatchChan
	eventChan chan *registry.Event
	client    *etcd_client.Client
}

// Next is a blocking call
func (w *_Watcher) Next() (*registry.Event, error) {
	for event := range w.eventChan {
		return event, nil
	}
	return nil, registry.ErrStoppedWatching
}

// Stop stop watching
func (w *_Watcher) Stop() {
	w.cancel()
}
