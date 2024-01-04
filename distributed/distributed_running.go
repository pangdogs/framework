package distributed

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"kit.golaxy.org/plugins/broker"
	"kit.golaxy.org/plugins/log"
	"kit.golaxy.org/plugins/registry"
	"time"
)

func (d *_Distributed) mainLoop(serviceNode registry.Service, subs []broker.Subscriber) {
	defer d.wg.Done()

	log.Infof(d.ctx, "service %q node %q started", d.ctx.GetName(), d.ctx.GetId())

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

			log.Debugf(d.ctx, "refresh service %q node %q success", d.ctx.GetName(), d.ctx.GetId())

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

	log.Infof(d.ctx, "service %q node %q stopped", d.ctx.GetName(), d.ctx.GetId())
}

func (d *_Distributed) watchingService(watcher registry.Watcher) {
	defer d.wg.Done()

	log.Debug(d.ctx, "watching service changes started")

loop:
	for {
		e, err := watcher.Next()
		if err != nil {
			if errors.Is(err, registry.ErrStoppedWatching) {
				break loop
			}
			log.Errorf(d.ctx, "watching service changes failed, %s", err)
			continue
		}

		switch e.Type {
		case registry.Delete:
			for _, node := range e.Service.Nodes {
				d.deduplication.Remove(node.Address)
			}
		}
	}

	// 停止监听服务节点
	<-watcher.Stop()

	log.Debug(d.ctx, "watching service changes stopped")
}

func (d *_Distributed) handleEvent(e broker.Event) error {
	mp, err := d.decoder.DecodeBytes(e.Message())
	if err != nil {
		return err
	}

	if !d.deduplication.ValidateSeq(mp.Head.Src, mp.Head.Seq) {
		return fmt.Errorf("gap: discard duplicate msg-packet, head:%+v", mp.Head)
	}

	var errs []error

	interrupt := func(err, _ error) bool {
		if err != nil {
			errs = append(errs, err)
		}
		return false
	}

	d.msgWatchers.AutoRLock(func(watchers *[]*_MsgWatcher) {
		for i := range *watchers {
			(*watchers)[i].handler.Exec(interrupt, e.Topic(), mp)
		}
	})

	d.options.RecvMsgHandler.Exec(interrupt, e.Topic(), mp)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
