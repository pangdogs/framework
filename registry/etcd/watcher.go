package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/registry"
	"time"
)

type _EtcdWatcher struct {
	ctx       service.Context
	stopChan  chan bool
	watchChan clientv3.WatchChan
	client    *clientv3.Client
	timeout   time.Duration
}

func newEtcdWatcher(ctx context.Context, r *_EtcdRegistry, timeout time.Duration, serviceName string) (registry.Watcher, error) {
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

	return &_EtcdWatcher{
		ctx:       r.ctx,
		stopChan:  stop,
		watchChan: r.client.Watch(ctx, watchPath, clientv3.WithPrefix(), clientv3.WithPrevKV()),
		client:    r.client,
		timeout:   timeout,
	}, nil
}

// Next is a blocking call
func (ew *_EtcdWatcher) Next() (*registry.Result, error) {
	for watchRsp := range ew.watchChan {
		if watchRsp.Err() != nil {
			return nil, watchRsp.Err()
		}
		if watchRsp.Canceled {
			return nil, registry.ErrWatcherStopped
		}

		for _, event := range watchRsp.Events {
			var service *registry.Service
			var action string
			var err error

			switch event.Type {
			case clientv3.EventTypePut:
				if event.IsCreate() {
					action = registry.Create.String()
				} else if event.IsModify() {
					action = registry.Update.String()
				}

				// get service from Kv
				service, err = decodeService(event.Kv.Value)

			case clientv3.EventTypeDelete:
				action = registry.Delete.String()

				// get service from prevKv
				service, err = decodeService(event.PrevKv.Value)

			default:
				logger.Debugf(ew.ctx, "unknown event type %q", event.Type)
				continue
			}

			if err != nil {
				logger.Debug(ew.ctx, err)
				continue
			}

			return &registry.Result{
				Action:  action,
				Service: service,
			}, nil
		}
	}

	return nil, registry.ErrWatcherStopped
}

// Stop stop watching
func (ew *_EtcdWatcher) Stop() {
	select {
	case <-ew.stopChan:
		return
	default:
		close(ew.stopChan)
	}
}
