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

package dsvc

import (
	"context"
	"errors"
	"fmt"
	"git.golaxy.org/framework/addins/broker"
	"git.golaxy.org/framework/addins/discovery"
	"git.golaxy.org/framework/addins/log"
	"time"
)

func (d *_DistService) mainLoop(serviceNode *discovery.Service, subs []broker.ISubscriber) {
	defer d.wg.Done()

	log.Infof(d.svcCtx, "service %q node %q started, localAddr:%q", d.svcCtx.GetName(), d.svcCtx.GetId(), d.details.LocalAddr)

	if d.options.RefreshTTL {
		ticker := time.NewTicker(d.options.TTL / 2)
		defer ticker.Stop()

	loop:
		for {
			select {
			case <-ticker.C:
				// 刷新服务节点
				if err := d.registry.Register(d.ctx, serviceNode, d.options.TTL); err != nil {
					log.Errorf(d.svcCtx, "refresh service %q node %q failed, %s", d.svcCtx.GetName(), d.svcCtx.GetId(), err)
					continue
				}

				log.Debugf(d.svcCtx, "refresh service %q node %q success", d.svcCtx.GetName(), d.svcCtx.GetId())

			case <-d.ctx.Done():
				break loop
			}
		}
	} else {
		<-d.ctx.Done()
	}

	// 取消注册服务节点
	if err := d.registry.Deregister(context.Background(), serviceNode); err != nil {
		log.Errorf(d.svcCtx, "deregister service %q node %q failed, %s", d.svcCtx.GetName(), d.svcCtx.GetId(), err)
	}

	// 取消订阅topic
	for _, sub := range subs {
		<-sub.Unsubscribe()
	}

	d.broker.Flush(context.Background())

	log.Infof(d.svcCtx, "service %q node %q terminated", d.svcCtx.GetName(), d.svcCtx.GetId())
}

func (d *_DistService) watchingService() {
	defer d.wg.Done()

	log.Debug(d.svcCtx, "watching service changes started")

retry:
	var watcher discovery.IWatcher
	var err error
	retryInterval := 3 * time.Second

	select {
	case <-d.ctx.Done():
		goto end
	default:
	}

	// 监控服务节点变化
	watcher, err = d.registry.Watch(d.ctx, "")
	if err != nil {
		log.Errorf(d.svcCtx, "watching service changes failed, %s, retry it", err)
		time.Sleep(retryInterval)
		goto retry
	}

	for {
		e, err := watcher.Next()
		if err != nil {
			if errors.Is(err, discovery.ErrTerminated) {
				time.Sleep(retryInterval)
				goto retry
			}

			log.Errorf(d.svcCtx, "watching service changes failed, %s, retry it", err)
			<-watcher.Terminate()
			time.Sleep(retryInterval)
			goto retry
		}

		switch e.Type {
		case discovery.Delete:
			for _, node := range e.Service.Nodes {
				d.deduplicator.Remove(node.Address)
			}
		}
	}

end:
	if watcher != nil {
		<-watcher.Terminate()
	}

	log.Debug(d.svcCtx, "watching service changes stopped")
}

func (d *_DistService) handleEvent(e broker.Event) error {
	mp, err := d.decoder.Decode(e.Message)
	if err != nil {
		return err
	}

	// 最少一次交付模式，需要消息去重
	if d.broker.GetDeliveryReliability() == broker.AtLeastOnce {
		if !d.deduplicator.Validate(mp.Head.Src.Addr, mp.Head.Seq) {
			return fmt.Errorf("dsvc: discard duplicate msg-packet, head:%+v", mp.Head)
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
			(*watchers)[i].handler.UnsafeCall(interrupt, e.Topic, mp)
		}
	})

	// 回调处理器
	d.options.RecvMsgHandler.UnsafeCall(interrupt, e.Topic, mp)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
