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

package dserv

import (
	"context"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"git.golaxy.org/framework/net/netpath"
	"git.golaxy.org/framework/plugins/broker"
	"git.golaxy.org/framework/plugins/discovery"
	"git.golaxy.org/framework/plugins/dsync"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/utils/concurrent"
	"github.com/josharian/intern"
	"sync"
)

// IWatcher 监听器
type IWatcher interface {
	context.Context
	Terminate() <-chan struct{}
	Terminated() <-chan struct{}
}

// IDistService 分布式服务支持
type IDistService interface {
	// GetNodeDetails 获取节点地址信息
	GetNodeDetails() *NodeDetails
	// GetFutures 获取异步模型Future控制器
	GetFutures() concurrent.IFutures
	// SendMsg 发送消息
	SendMsg(dst string, msg gap.Msg) error
	// ForwardMsg 转发消息
	ForwardMsg(svc, src, dst string, seq int64, msg gap.Msg) error
	// WatchMsg 监听消息（优先级高）
	WatchMsg(ctx context.Context, handler RecvMsgHandler) IWatcher
}

func newDistService(setting ...option.Setting[DistServiceOptions]) IDistService {
	return &_DistService{
		options: option.Make(With.Default(), setting...),
	}
}

type _DistService struct {
	servCtx       service.Context
	ctx           context.Context
	terminate     context.CancelFunc
	wg            sync.WaitGroup
	options       DistServiceOptions
	registry      discovery.IRegistry
	broker        broker.IBroker
	dsync         dsync.IDistSync
	details       *NodeDetails
	encoder       codec.Encoder
	decoder       codec.Decoder
	futures       concurrent.Futures
	sendMutex     sync.Mutex
	deduplication concurrent.Deduplication
	msgWatchers   concurrent.LockedSlice[*_MsgWatcher]
}

// InitSP 初始化服务插件
func (d *_DistService) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", self.Name)

	d.servCtx = ctx
	d.ctx, d.terminate = context.WithCancel(context.Background())

	// 获取依赖的插件
	d.registry = discovery.Using(d.servCtx)
	d.broker = broker.Using(d.servCtx)
	d.dsync = dsync.Using(d.servCtx)

	// 初始化消息包编解码器
	d.decoder = codec.MakeDecoder(d.options.DecoderMsgCreator)
	d.encoder = codec.MakeEncoder()

	// 初始化异步模型Future
	d.futures = concurrent.MakeFutures(d.ctx, d.options.FutureTimeout)

	// 初始化消息去重器
	d.deduplication = concurrent.MakeDeduplication()

	// 初始化监听器
	d.msgWatchers = concurrent.MakeLockedSlice[*_MsgWatcher](0, 0)

	// 初始化地址信息
	details := &NodeDetails{}
	sep := d.broker.GetSeparator()

	details.DomainRoot = netpath.Domain{
		Path: intern.String(d.options.DomainRoot),
		Sep:  sep,
	}
	details.DomainBroadcast = netpath.Domain{
		Path: intern.String(netpath.Join(sep, details.DomainRoot.Path, "bc")),
		Sep:  sep,
	}
	details.DomainBalance = netpath.Domain{
		Path: intern.String(netpath.Join(sep, details.DomainRoot.Path, "lb")),
		Sep:  sep,
	}
	details.DomainUnicast = netpath.Domain{
		Path: intern.String(netpath.Join(sep, details.DomainRoot.Path, "nd")),
		Sep:  sep,
	}

	details.GlobalBroadcastAddr = details.DomainBroadcast.Path
	details.GlobalBalanceAddr = details.DomainBalance.Path
	details.BroadcastAddr = details.MakeBroadcastAddr(d.servCtx.GetName())
	details.BalanceAddr = details.MakeBalanceAddr(d.servCtx.GetName())
	details.LocalAddr, _ = details.MakeNodeAddr(d.servCtx.GetId())

	d.details = details
	log.Debugf(d.servCtx, "service %q node %q details: %+v", d.servCtx.GetName(), d.servCtx.GetId(), d.details)

	// 加分布式锁
	mutex := d.dsync.NewMutex(netpath.Join(d.dsync.GetSeparator(), "service", d.servCtx.GetName(), "init", d.servCtx.GetId().String()))
	if err := mutex.Lock(d.servCtx); err != nil {
		log.Panicf(d.servCtx, "lock dsync mutex %q failed, %s", mutex.Name(), err)
	}
	defer mutex.Unlock(context.Background())

	// 检查服务节点是否已被注册
	_, err := d.registry.GetServiceNode(d.servCtx, d.servCtx.GetName(), d.servCtx.GetId())
	if err == nil {
		log.Panicf(d.servCtx, "check service %q node %q failed, already registered", d.servCtx.GetName(), d.servCtx.GetId())
	}
	if !errors.Is(err, discovery.ErrNotFound) {
		log.Panicf(d.servCtx, "check service %q node %q failed, %s", d.servCtx.GetName(), d.servCtx.GetId(), err)
	}

	// 订阅topic
	subs := []broker.ISubscriber{
		// 订阅全服topic
		d.subscribe(d.details.GlobalBroadcastAddr, ""),
		d.subscribe(d.details.GlobalBalanceAddr, "balance"),
		// 订阅服务类型topic
		d.subscribe(d.details.BroadcastAddr, ""),
		d.subscribe(d.details.BalanceAddr, "balance"),
		// 订阅服务节点topic
		d.subscribe(d.details.LocalAddr, ""),
	}

	// 服务节点信息
	serviceNode := &discovery.Service{
		Name: d.servCtx.GetName(),
		Nodes: []discovery.Node{
			{
				Id:      d.servCtx.GetId(),
				Address: d.details.LocalAddr,
				Version: d.options.Version,
				Meta:    d.options.Meta,
			},
		},
	}

	// 注册服务
	err = d.registry.Register(d.servCtx, serviceNode, d.options.TTL)
	if err != nil {
		log.Panicf(d.servCtx, "register service %q node %q failed, %s", d.servCtx.GetName(), d.servCtx.GetId(), err)
	}
	log.Debugf(d.servCtx, "register service %q node %q success", d.servCtx.GetName(), d.servCtx.GetId())

	// 最少一次交付模式，需要消息去重
	if d.broker.GetDeliveryReliability() == broker.AtLeastOnce {
		// 运行服务节点监听线程
		d.wg.Add(1)
		go d.watchingService()
	}

	// 运行主线程
	d.wg.Add(1)
	go d.mainLoop(serviceNode, subs)
}

// ShutSP 关闭服务插件
func (d *_DistService) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", self.Name)

	d.terminate()
	d.wg.Wait()
}

// GetNodeDetails 获取节点地址信息
func (d *_DistService) GetNodeDetails() *NodeDetails {
	return d.details
}

// GetFutures 获取异步模型Future控制器
func (d *_DistService) GetFutures() concurrent.IFutures {
	return &d.futures
}

// SendMsg 发送消息
func (d *_DistService) SendMsg(dst string, msg gap.Msg) error {
	if msg == nil {
		return fmt.Errorf("%w: msg is nil", core.ErrArgs)
	}

	var seq int64

	// 最少一次交付模式，需要消息去重
	if d.broker.GetDeliveryReliability() == broker.AtLeastOnce {
		d.sendMutex.Lock()
		defer d.sendMutex.Unlock()
		seq = d.deduplication.Make()
	}

	mpBuf, err := d.encoder.Encode(d.servCtx.GetName(), d.details.LocalAddr, seq, msg)
	if err != nil {
		return err
	}
	defer mpBuf.Release()

	return d.broker.Publish(d.ctx, dst, mpBuf.Data())
}

// ForwardMsg 转发消息
func (d *_DistService) ForwardMsg(svc, src, dst string, seq int64, msg gap.Msg) error {
	if msg == nil {
		return fmt.Errorf("%w: msg is nil", core.ErrArgs)
	}

	mpBuf, err := d.encoder.Encode(svc, src, seq, msg)
	if err != nil {
		return err
	}
	defer mpBuf.Release()

	return d.broker.Publish(d.ctx, dst, mpBuf.Data())
}

// WatchMsg 监听消息（优先级高）
func (d *_DistService) WatchMsg(ctx context.Context, handler RecvMsgHandler) IWatcher {
	return d.newMsgWatcher(ctx, handler)
}

func (d *_DistService) subscribe(topic, queue string) broker.ISubscriber {
	sub, err := d.broker.Subscribe(d.ctx, topic,
		broker.With.EventHandler(generic.MakeDelegateFunc1(d.handleEvent)),
		broker.With.Queue(queue))
	if err != nil {
		log.Panicf(d.servCtx, "subscribe topic %q queue %q failed, %s", topic, queue, err)
	}
	log.Debugf(d.servCtx, "subscribe topic %q queue %q success", topic, queue)
	return sub
}
