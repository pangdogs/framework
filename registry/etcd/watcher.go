package etcd

import (
	"context"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"kit.golaxy.org/plugins/registry"
	"time"
)

type _EtcdWatcher struct {
	stopChan  chan bool
	watchChan clientv3.WatchChan
	client    *clientv3.Client
	timeout   time.Duration
}

func newEtcdWatcher(ctx context.Context, r *_EtcdRegistry, timeout time.Duration, serviceName string) (registry.Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
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
		stopChan:  stop,
		watchChan: r.client.Watch(ctx, watchPath, clientv3.WithPrefix(), clientv3.WithPrevKV()),
		client:    r.client,
		timeout:   timeout,
	}, nil
}

func (ew *_EtcdWatcher) Next() (*registry.Result, error) {
	for watchRsp := range ew.watchChan {
		if watchRsp.Err() != nil {
			return nil, watchRsp.Err()
		}
		if watchRsp.Canceled {
			return nil, errors.New("could not get next")
		}

		for _, ev := range watchRsp.Events {
			var service *registry.Service
			var action string

			switch ev.Type {
			case clientv3.EventTypePut:
				if ev.IsCreate() {
					action = registry.Create.String()
				} else if ev.IsModify() {
					action = registry.Update.String()
				}

				// get service from Kv
				service = decodeService(ev.Kv.Value)

			case clientv3.EventTypeDelete:
				action = registry.Delete.String()

				// get service from prevKv
				service = decodeService(ev.PrevKv.Value)
			}

			if service == nil || action == "" {
				continue
			}

			return &registry.Result{
				Action:  action,
				Service: service,
			}, nil
		}
	}

	return nil, errors.New("could not get next")
}

func (ew *_EtcdWatcher) Stop() {
	select {
	case <-ew.stopChan:
		return
	default:
		close(ew.stopChan)
	}
}
