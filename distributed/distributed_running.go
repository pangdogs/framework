package distributed

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/plugins/broker"
	"kit.golaxy.org/plugins/log"
	"kit.golaxy.org/plugins/registry"
	"time"
)

func (d *_Distributed) mainLoop(serviceNode registry.Service, subs []broker.Subscriber) {
	defer d.wg.Done()

	log.Infof(d.ctx, "start service %q node %q", d.ctx.GetName(), d.ctx.GetId())

	ticker := time.NewTicker(d.options.RefreshInterval)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			// 刷新服务节点
			if err := d.registry.Register(d.ctx, serviceNode, d.options.RefreshInterval*2); err != nil {
				log.Errorf(d.ctx, "refresh service %q node %q failed, %s", d.ctx.GetName(), d.ctx.GetId(), err)
				continue
			}

			log.Infof(d.ctx, "refresh service %q node %q success", d.ctx.GetName(), d.ctx.GetId())

		case <-d.ctx.Done():
			break loop
		}
	}

	// 取消注册服务节点
	if err := d.registry.Deregister(context.Background(), serviceNode); err != nil {
		log.Errorf(d.ctx, "deregister service %q node %q failed, %s", d.ctx.GetName(), d.ctx.GetId(), err)
	}

	// 取消订阅topic
	for _, sub := range subs {
		<-sub.Unsubscribe()
	}

	log.Infof(d.ctx, "stop service %q node %q", d.ctx.GetName(), d.ctx.GetId())
}

func (d *_Distributed) handleEvent(e broker.Event) error {
	mp, err := d.decoder.DecodeBytes(e.Message())
	if err != nil {
		return err
	}

	if !d.deduplication.ValidateSeq(mp.Head.Src, mp.Head.Seq) {
		return fmt.Errorf("gap: discard duplicate msg-packet, head:%+v", mp.Head)
	}

	return generic.FuncError(d.options.RecvMsgHandler.Invoke(nil, e.Topic(), mp))
}

func (d *_Distributed) watching(watcher registry.Watcher) {
	defer d.wg.Done()

loop:
	for {
		e, err := watcher.Next()
		if err != nil {
			if errors.Is(err, registry.ErrStoppedWatching) {
				log.Debugf(d.ctx, "watching service changes stopped")
				break loop
			}
			log.Errorf(d.ctx, "watching service changes aborted, %s", err)
			break loop
		}

		switch e.Type {
		case registry.Delete:
			for _, node := range e.Service.Nodes {
				d.deduplication.Remove(node.Address)
			}
		}
	}

	<-watcher.Stop()
}
