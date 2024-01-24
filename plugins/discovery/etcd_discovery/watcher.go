package etcd_discovery

import (
	"context"
	"git.golaxy.org/framework/plugins/discovery"
	"git.golaxy.org/framework/plugins/log"
	etcd_client "go.etcd.io/etcd/client/v3"
	"strings"
)

func (r *_Registry) newWatcher(ctx context.Context, pattern string) (*_Watcher, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	watchKey := r.options.KeyPrefix
	if pattern != "" {
		watchKey = getServicePath(r.options.KeyPrefix, pattern)
	}

	watcher := &_Watcher{
		registry:    r,
		ctx:         ctx,
		cancel:      cancel,
		stoppedChan: make(chan struct{}),
		pattern:     pattern,
		watchChan:   r.client.Watch(ctx, watchKey, etcd_client.WithPrefix(), etcd_client.WithPrevKV()),
		eventChan:   make(chan *discovery.Event, r.options.WatchChanSize),
	}

	go watcher.mainLoop()

	return watcher, nil
}

type _Watcher struct {
	registry    *_Registry
	ctx         context.Context
	cancel      context.CancelFunc
	stoppedChan chan struct{}
	pattern     string
	watchChan   etcd_client.WatchChan
	eventChan   chan *discovery.Event
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
	return nil, discovery.ErrStoppedWatching
}

// Stop stop watching
func (w *_Watcher) Stop() <-chan struct{} {
	w.cancel()
	return w.stoppedChan
}

func (w *_Watcher) mainLoop() {
	defer func() {
		w.cancel()
		close(w.eventChan)
		close(w.stoppedChan)
	}()

	log.Debugf(w.registry.servCtx, "start watch %q", w.pattern)

	for watchRsp := range w.watchChan {
		if watchRsp.Canceled {
			log.Debugf(w.registry.servCtx, "stop watch %q", w.pattern)
			return
		}
		if watchRsp.Err() != nil {
			log.Errorf(w.registry.servCtx, "interrupt watch %q, %s", w.pattern, watchRsp.Err())
			return
		}

		for _, etcdEvent := range watchRsp.Events {
			event := &discovery.Event{}
			var err error

			switch etcdEvent.Type {
			case etcd_client.EventTypePut:
				if etcdEvent.IsCreate() {
					event.Type = discovery.Create
				} else if etcdEvent.IsModify() {
					event.Type = discovery.Update
				}

				// get service from Kv
				event.Service, err = decodeService(etcdEvent.Kv.Value)
				if err != nil {
					log.Errorf(w.registry.servCtx, "decode service %q failed, %s", etcdEvent.Kv.Value, err)
					continue
				}

			case etcd_client.EventTypeDelete:
				event.Type = discovery.Delete

				// get service from prevKv
				event.Service, err = decodeService(etcdEvent.PrevKv.Value)
				if err != nil {
					log.Errorf(w.registry.servCtx, "decode service %q failed, %s", etcdEvent.PrevKv.Value, err)
					continue
				}

			default:
				log.Errorf(w.registry.servCtx, "unknown event type %q", etcdEvent.Type)
				continue
			}

			if len(event.Service.Nodes) <= 0 {
				log.Warnf(w.registry.servCtx, "event service %q node is empty, discard it", event.Service.Name)
				continue
			}

			select {
			case w.eventChan <- event:
			case <-w.ctx.Done():
				log.Debugf(w.registry.servCtx, "stop watch %q", w.pattern)
				return
			}
		}
	}
}
