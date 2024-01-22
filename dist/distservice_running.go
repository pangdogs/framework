package dist

import (
	"context"
	"errors"
	"fmt"
	"git.golaxy.org/plugins/broker"
	"git.golaxy.org/plugins/discovery"
	"git.golaxy.org/plugins/log"
	"time"
)

func (d *_DistService) mainLoop(serviceNode *discovery.Service, subs []broker.ISubscriber) {
	defer d.wg.Done()

	log.Infof(d.servCtx, "service %q node %q started", d.servCtx.GetName(), d.servCtx.GetId())

	ticker := time.NewTicker(d.options.RefreshInterval)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			// 刷新服务节点
			if err := d.registry.Register(d.ctx, serviceNode, d.options.RefreshInterval*2); err != nil {
				log.Errorf(d.servCtx, "refresh service %q node %q failed, %s", d.servCtx.GetName(), d.servCtx.GetId(), err)
				continue
			}

			log.Debugf(d.servCtx, "refresh service %q node %q success", d.servCtx.GetName(), d.servCtx.GetId())

		case <-d.ctx.Done():
			break loop
		}
	}

	// 取消注册服务节点
	if err := d.registry.Deregister(context.Background(), serviceNode); err != nil {
		log.Errorf(d.servCtx, "deregister service %q node %q failed, %s", d.servCtx.GetName(), d.servCtx.GetId(), err)
	}

	// 取消订阅topic
	for _, sub := range subs {
		<-sub.Unsubscribe()
	}

	d.broker.Flush(context.Background())

	log.Infof(d.servCtx, "service %q node %q stopped", d.servCtx.GetName(), d.servCtx.GetId())
}

func (d *_DistService) watchingService(watcher discovery.IWatcher) {
	defer d.wg.Done()

	log.Debug(d.servCtx, "watching service changes started")

loop:
	for {
		e, err := watcher.Next()
		if err != nil {
			if errors.Is(err, discovery.ErrStoppedWatching) {
				break loop
			}
			log.Errorf(d.servCtx, "watching service changes failed, %s", err)
			continue
		}

		switch e.Type {
		case discovery.Delete:
			for _, node := range e.Service.Nodes {
				d.deduplication.Remove(node.Address)
			}
		}
	}

	// 停止监听服务节点
	<-watcher.Stop()

	log.Debug(d.servCtx, "watching service changes stopped")
}

func (d *_DistService) handleEvent(e broker.IEvent) error {
	mp, err := d.decoder.DecodeBytes(e.Message())
	if err != nil {
		return err
	}

	// 最少一次交付模式，需要消息去重
	if d.broker.GetDeliveryReliability() == broker.AtLeastOnce {
		if !d.deduplication.ValidateSeq(mp.Head.Src, mp.Head.Seq) {
			return fmt.Errorf("gap: discard duplicate msg-packet, head:%+v", mp.Head)
		}
	}

	var errs []error

	interrupt := func(err, _ error) bool {
		if err != nil {
			errs = append(errs, err)
		}
		return false
	}

	// 回调监控器
	d.msgWatchers.AutoRLock(func(watchers *[]*_MsgWatcher) {
		for i := range *watchers {
			(*watchers)[i].handler.Exec(interrupt, e.Topic(), mp)
		}
	})

	// 回调处理器
	d.options.RecvMsgHandler.Exec(interrupt, e.Topic(), mp)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
