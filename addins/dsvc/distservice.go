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
	"sync"
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

// IDistService 提供服务节点发布、GAP 消息收发及请求 Future 管理。
type IDistService interface {
	// BringUp 订阅节点地址并将当前节点注册到服务发现；重复调用不会重复上线。
	BringUp()
	// NodeDetails 返回当前服务节点的地址信息；调用方不得修改。
	NodeDetails() *NodeDetails
	// FutureController 返回用于关联请求与响应的 Future 控制器。
	FutureController() *concurrent.FutureController
	// Send 编码 msg 并发布到 dst。
	Send(dst string, msg gap.Msg) error
	// Listen 注册消息处理器，直到 ctx 取消或 add-in 停止。
	// 返回的 Signal 在监听器移除后完成。
	Listen(ctx context.Context, handler MsgHandler) (async.Signal, error)
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
	bringUpOnce      sync.Once
	listeners        concurrent.Listeners[MsgHandler, _BrokerMsg]
}

// Init 获取服务发现、消息中间件和分布式同步依赖，并创建编解码器、Future 控制器及节点地址。
func (d *_DistService) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	d.svcCtx = svcCtx
	d.ctx, d.terminate = context.WithCancel(context.Background())

	// 获取上线及消息收发所需的 add-in。
	d.registry = discovery.AddIn.Require(svcCtx)
	d.broker = broker.AddIn.Require(svcCtx)
	d.dsync = dsync.AddIn.Require(svcCtx)

	// 当前分布式服务处理链仅支持至多一次交付。
	if d.broker.DeliveryReliability() != broker.DeliveryReliability_AtMostOnce {
		log.L(svcCtx).Panic("broker delivery reliability must be at most once")
	}

	// 初始化 GAP 消息包编解码器。
	d.decoder = codec.NewDecoder(d.options.MsgCreator)
	d.encoder = codec.NewEncoder()

	// Future 控制器随 add-in 的内部上下文一起停止。
	d.futureController = concurrent.NewFutureController(d.ctx, d.options.FutureTimeout)

	// 根据 broker 分隔符和服务身份生成各级消息地址。
	d.initNodeDetails()
}

// Shut 取消内部上下文，拒绝新任务，并等待节点注销、消息退订及监听器退出。
func (d *_DistService) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))

	d.terminate()
	d.barrier.Close()
	d.barrier.Wait()
}

// BringUp 仅执行一次：先订阅节点地址，再通过分布式锁检查并注册当前服务节点。
// add-in 停止时会注销节点、等待订阅退出并刷新 broker。
func (d *_DistService) BringUp() {
	d.bringUpOnce.Do(func() {
		svcCtx := d.svcCtx

		if !d.barrier.Join(1) {
			log.L(svcCtx).Panic("service node is terminating")
		}

		log.L(svcCtx).Info("service node is starting",
			zap.String("service", svcCtx.Name()),
			zap.String("node", svcCtx.Id().String()),
			log.JSON("details", d.details))

		// 在注册节点前订阅全部五类接收地址，避免上线后遗漏消息。
		subs := []async.Signal{
			// 全局广播与全局负载均衡地址。
			d.subscribe(d.details.GlobalBroadcastAddr, ""),
			d.subscribe(d.details.GlobalBalanceAddr, "balance"),

			// 当前服务类型的广播与负载均衡地址。
			d.subscribe(d.details.BroadcastAddr, ""),
			d.subscribe(d.details.BalanceAddr, "balance"),

			// 当前节点的单播地址。
			d.subscribe(d.details.LocalAddr, ""),
		}

		// 串行化同名、同 ID 节点的查重与注册。
		mutex := d.dsync.NewMutex(netpath.Join(d.dsync.Separator(), "service_node_start", svcCtx.Name(), svcCtx.Id().String()))
		if err := mutex.Lock(svcCtx); err != nil {
			log.L(svcCtx).Panic("lock dsync mutex failed", zap.String("name", mutex.Name()), zap.Error(err))
		}
		defer mutex.Unlock(context.Background())

		// 已存在的同名节点视为配置冲突，不接管其租约。
		_, err := d.registry.GetNode(svcCtx, svcCtx.Name(), svcCtx.Id())
		if err == nil {
			log.L(svcCtx).Panic("service node already registered", zap.String("service", svcCtx.Name()), zap.String("node", svcCtx.Id().String()))
		}
		if !errors.Is(err, discovery.ErrRegistrationNotFound) {
			log.L(svcCtx).Panic("checking service node failed", zap.String("service", svcCtx.Name()), zap.String("node", svcCtx.Id().String()), zap.Error(err))
		}

		// 发布节点单播地址及调用方配置的版本和元数据。
		node := &discovery.Node{
			Id:      svcCtx.Id(),
			Address: d.details.LocalAddr,
			Version: d.options.Version,
			Meta:    d.options.Meta,
		}

		// 注册后持续续租，直到 add-in 的内部上下文取消。
		reg, err := d.registry.RegisterNode(d.ctx, svcCtx.Name(), node, d.options.RegistrationTTL)
		if err != nil {
			log.L(svcCtx).Panic("register service node failed",
				zap.String("service", svcCtx.Name()),
				zap.String("node", svcCtx.Id().String()),
				zap.Error(err))
		}
		if _, err = reg.KeepAliveContinuous(d.ctx); err != nil {
			log.L(svcCtx).Panic("keepalive service node failed",
				zap.String("service", svcCtx.Name()),
				zap.String("node", svcCtx.Id().String()),
				zap.Error(err))
		}

		log.L(svcCtx).Info("service node is started",
			zap.String("service", svcCtx.Name()),
			zap.String("node", svcCtx.Id().String()),
			log.JSON("details", d.details))

		go func() {
			defer d.barrier.Done()
			<-d.ctx.Done()
			// 停止时先注销节点，使发现端不再把流量路由到本节点。
			reg.Deregister(context.Background())
			// 各订阅随 d.ctx 取消；等待全部退订完成后再刷新 broker。
			for _, sub := range subs {
				<-sub.Done()
			}
			d.broker.Flush(context.Background())
		}()
	})
}

// NodeDetails 返回当前服务节点的地址信息；调用方不得修改。
func (d *_DistService) NodeDetails() *NodeDetails {
	return d.details
}

// FutureController 返回用于关联请求与响应的 Future 控制器。
func (d *_DistService) FutureController() *concurrent.FutureController {
	return d.futureController
}

// Send 将 msg 编码为 GAP 消息包并发布到 dst。
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

// Listen 注册消息处理器，直到 ctx 取消或 add-in 停止；handler 为 nil 时返回错误。
func (d *_DistService) Listen(ctx context.Context, handler MsgHandler) (async.Signal, error) {
	if handler == nil {
		return async.Signal{}, errors.New("dsvc: handler is nil")
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

func (d *_DistService) subscribe(topic, queue string) async.Signal {
	unsubscribed, err := d.broker.SubscribeHandler(d.ctx, topic, queue, generic.CastDelegateVoid1(d.handleEvent))
	if err != nil {
		log.L(d.svcCtx).Panic("subscribe service broker event failed", zap.String("topic", topic), zap.String("queue", queue), zap.Error(err))
	}
	log.L(d.svcCtx).Info("subscribe service broker event ok", zap.String("topic", topic), zap.String("queue", queue))
	return unsubscribed
}
