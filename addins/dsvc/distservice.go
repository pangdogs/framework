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
	"time"
	"unique"

	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/broker"
	"git.golaxy.org/framework/addins/discovery"
	"git.golaxy.org/framework/addins/dsync"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"git.golaxy.org/framework/net/netpath"
	"git.golaxy.org/framework/utils/concurrent"
	"go.uber.org/zap"
)

// IDistService 分布式服务支持
type IDistService interface {
	// NodeDetails 获取节点地址信息
	NodeDetails() *NodeDetails
	// FutureController 获取异步模型Future控制器
	FutureController() *concurrent.FutureController
	// Send 发送消息
	Send(dst string, msg gap.Msg) error
	// Listen 监听消息
	Listen(ctx context.Context, handler MsgHandler) error
}

func newDistService(setting ...option.Setting[DistServiceOptions]) IDistService {
	return &_DistService{
		options: option.New(With.Default(), setting...),
	}
}

type _DistService struct {
	svcCtx           service.Context
	ctx              context.Context
	terminate        context.CancelFunc
	barrier          generic.Barrier
	options          DistServiceOptions
	registry         discovery.IRegistry
	broker           broker.IBroker
	dsync            dsync.IDistSync
	details          *NodeDetails
	encoder          *codec.Encoder
	decoder          *codec.Decoder
	futureController *concurrent.FutureController
	listeners        concurrent.Listeners[MsgHandler, _BrokerMsg]
}

// Init 初始化插件
func (d *_DistService) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	d.svcCtx = svcCtx
	d.ctx, d.terminate = context.WithCancel(context.Background())

	// 获取依赖的插件
	d.registry = discovery.AddIn.Require(svcCtx)
	d.broker = broker.AddIn.Require(svcCtx)
	d.dsync = dsync.AddIn.Require(svcCtx)

	// 检测broker的交付模式
	if d.broker.DeliveryReliability() != broker.DeliveryReliability_AtMostOnce {
		log.L(svcCtx).Panic("broker delivery reliability must be at most once")
	}

	// 初始化消息包编解码器
	d.decoder = codec.NewDecoder(d.options.MsgCreator)
	d.encoder = codec.NewEncoder()

	// 初始化异步模型Future控制器
	d.futureController = concurrent.NewFutureController(d.ctx, d.options.FutureTimeout)

	// 初始化地址信息
	d.initNodeDetails()

	log.L(svcCtx).Info("service node is starting",
		zap.String("service", svcCtx.Name()),
		zap.String("node", svcCtx.Id().String()),
		log.JSON("details", d.details))

	// 加分布式锁
	mutex := d.dsync.NewMutex(netpath.Join(d.dsync.Separator(), "service_node_start", svcCtx.Name(), svcCtx.Id().String()))
	if err := mutex.Lock(svcCtx); err != nil {
		log.L(svcCtx).Panic("lock dsync mutex failed", zap.String("name", mutex.Name()), zap.Error(err))
	}
	defer mutex.Unlock(context.Background())

	// 检查服务节点是否已被注册
	_, err := d.registry.GetNode(svcCtx, svcCtx.Name(), svcCtx.Id())
	if err == nil {
		log.L(svcCtx).Panic("service node already registered", zap.String("service", svcCtx.Name()), zap.String("node", svcCtx.Id().String()))
	}
	if !errors.Is(err, discovery.ErrRegistrationNotFound) {
		log.L(svcCtx).Panic("checking service node failed", zap.String("service", svcCtx.Name()), zap.String("node", svcCtx.Id().String()), zap.Error(err))
	}

	// 订阅消息事件
	subs := []async.Future{
		// 订阅全服消息事件
		d.subscribe(d.details.GlobalBroadcastAddr, ""),
		d.subscribe(d.details.GlobalBalanceAddr, "balance"),

		// 订阅服务类型消息事件
		d.subscribe(d.details.BroadcastAddr, ""),
		d.subscribe(d.details.BalanceAddr, "balance"),

		// 订阅服务节点消息事件
		d.subscribe(d.details.LocalAddr, ""),
	}

	// 服务节点信息
	node := &discovery.Node{
		Id:      svcCtx.Id(),
		Address: d.details.LocalAddr,
		Version: d.options.Version,
		Meta:    d.options.Meta,
	}

	// 注册服务节点
	reg, err := d.registry.RegisterNode(svcCtx, svcCtx.Name(), node, d.options.RegistrationTTL, true)
	if err != nil {
		log.L(svcCtx).Panic("register service node failed",
			zap.String("service", svcCtx.Name()),
			zap.String("node", svcCtx.Id().String()),
			zap.Error(err))
	}

	log.L(svcCtx).Info("service node is started",
		zap.String("service", svcCtx.Name()),
		zap.String("node", svcCtx.Id().String()),
		log.JSON("details", d.details))

	d.barrier.Join(1)
	go func() {
		defer d.barrier.Done()
		<-d.ctx.Done()
		// 取消注册服务节点
		reg.Deregister(context.Background())
		// 等待消息事件已取消订阅
		for _, sub := range subs {
			<-sub.Done()
		}
		// 刷新消息中间件缓存
		d.broker.Flush(context.Background())
	}()
}

// Shut 关闭插件
func (d *_DistService) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))

	d.terminate()
	d.barrier.Close()
	d.barrier.Wait()
}

// NodeDetails 获取节点地址信息
func (d *_DistService) NodeDetails() *NodeDetails {
	return d.details
}

// FutureController 获取异步模型Future控制器
func (d *_DistService) FutureController() *concurrent.FutureController {
	return d.futureController
}

// Send 发送消息
func (d *_DistService) Send(dst string, msg gap.Msg) error {
	if msg == nil {
		return fmt.Errorf("dsvc: %w: msg is nil", core.ErrArgs)
	}

	mpBuf, err := d.encoder.Encode(
		gap.Origin{Svc: d.svcCtx.Name(), Addr: d.details.LocalAddr, Timestamp: time.Now().UnixMilli()},
		0,
		msg,
	)
	if err != nil {
		log.L(d.svcCtx).Error("encode message failed",
			zap.String("dst", dst),
			zap.Uint32("msg", msg.MsgId()),
			zap.Error(err))
		return fmt.Errorf("dsvc: %w", err)
	}
	defer mpBuf.Release()

	err = d.broker.Publish(d.ctx, dst, mpBuf.Payload())
	if err != nil {
		log.L(d.svcCtx).Error("publish message failed",
			zap.String("dst", dst),
			zap.Uint32("msg", msg.MsgId()),
			zap.Error(err))
		return fmt.Errorf("dsvc: %w", err)
	}

	return nil
}

// Listen 监听消息
func (d *_DistService) Listen(ctx context.Context, handler MsgHandler) error {
	if handler == nil {
		return errors.New("dsvc: handler is nil")
	}
	return d.addListener(ctx, handler)
}

func (d *_DistService) initNodeDetails() {
	details := &NodeDetails{}
	sep := d.broker.Separator()

	details.DomainRoot = netpath.Domain{
		Path: unique.Make(d.options.DomainRoot).Value(),
		Sep:  sep,
	}
	details.DomainBroadcast = netpath.Domain{
		Path: unique.Make(netpath.Join(sep, details.DomainRoot.Path, "bc")).Value(),
		Sep:  sep,
	}
	details.DomainBalance = netpath.Domain{
		Path: unique.Make(netpath.Join(sep, details.DomainRoot.Path, "lb")).Value(),
		Sep:  sep,
	}
	details.DomainUnicast = netpath.Domain{
		Path: unique.Make(netpath.Join(sep, details.DomainRoot.Path, "ep")).Value(),
		Sep:  sep,
	}

	details.GlobalBroadcastAddr = details.DomainBroadcast.Path
	details.GlobalBalanceAddr = details.DomainBalance.Path
	details.BroadcastAddr = details.MakeBroadcastAddr(d.svcCtx.Name())
	details.BalanceAddr = details.MakeBalanceAddr(d.svcCtx.Name())
	details.LocalAddr, _ = details.MakeNodeAddr(d.svcCtx.Id())

	d.details = details
}

func (d *_DistService) subscribe(topic, queue string) async.Future {
	unsubscribed, err := d.broker.SubscribeHandler(d.ctx, topic, queue, generic.CastDelegateVoid1(d.handleEvent))
	if err != nil {
		log.L(d.svcCtx).Panic("subscribe service broker event failed", zap.String("topic", topic), zap.String("queue", queue), zap.Error(err))
	}
	log.L(d.svcCtx).Info("subscribe service broker event ok", zap.String("topic", topic), zap.String("queue", queue))
	return unsubscribed
}
