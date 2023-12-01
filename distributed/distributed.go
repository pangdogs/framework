package distributed

import (
	"errors"
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/golaxy/util/option"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/broker"
	"kit.golaxy.org/plugins/dsync"
	"kit.golaxy.org/plugins/gap"
	"kit.golaxy.org/plugins/gap/codec"
	"kit.golaxy.org/plugins/log"
	"kit.golaxy.org/plugins/registry"
	"kit.golaxy.org/plugins/util/concurrent"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Distributed 分布式服务支持
type Distributed interface {
	// SendMsg 发送消息
	SendMsg(dst string, msg gap.Msg) error
	// GetFutures 获取异步模型Future控制器
	GetFutures() concurrent.IFutures
	// GetAddress 获取服务节点地址
	GetAddress() string
}

func newDistributed(setting ...option.Setting[DistributedOptions]) Distributed {
	return &_Distributed{
		Options: option.Make(Option{}.Default(), setting...),
	}
}

type _Distributed struct {
	ctx      service.Context
	Options  DistributedOptions
	registry registry.Registry
	broker   broker.Broker
	dsync    dsync.DSync
	service  registry.Service
	subs     []broker.Subscriber
	futures  concurrent.Futures
	sendSeq  int64
	recvSeq  map[string]int64
	wg       sync.WaitGroup
}

// InitSP 初始化服务插件
func (d *_Distributed) InitSP(ctx service.Context) {
	log.Infof(ctx, "init service plugin %q with %q", Name, types.AnyFullName(*d))

	d.ctx = ctx

	// 获取依赖的插件
	d.registry = registry.Using(ctx)
	d.broker = broker.Using(ctx)
	d.dsync = dsync.Using(ctx)

	// 异步模型Future
	d.futures.Ctx = d.ctx
	d.futures.Id = rand.Int63()
	d.futures.Timeout = d.Options.FutureTimeout

	// 初始化序号
	d.sendSeq = time.Now().UnixMicro()
	d.recvSeq = map[string]int64{}

	// 加分布式锁
	mutex := d.dsync.NewMutex(dsync.Path(d.ctx, "service", d.ctx.GetName(), d.ctx.GetId().String()))
	if err := mutex.Lock(d.ctx); err != nil {
		log.Panicf(d.ctx, "lock dsync mutex %q failed, %s", mutex.Name(), err)
	}
	defer mutex.Unlock(context.Background())

	// 检查服务节点是否已被注册
	_, err := d.registry.GetServiceNode(d.ctx, d.ctx.GetName(), d.ctx.GetId().String())
	if err == nil {
		log.Panicf(d.ctx, "check service %q node %q failed, already registered", d.ctx.GetName(), d.ctx.GetId())
	}
	if !errors.Is(err, registry.ErrNotFound) {
		log.Panicf(d.ctx, "check service %q node %q failed, %s", d.ctx.GetName(), d.ctx.GetId(), err)
	}

	// 服务信息
	d.service = registry.Service{
		Name: d.ctx.GetName(),
		Nodes: []registry.Node{
			{
				Id:      d.ctx.GetId().String(),
				Address: broker.Path(d.ctx, "service", d.ctx.GetName(), d.ctx.GetId().String()),
			},
		},
	}

	// 订阅topic
	d.subs = append(d.subs,
		// 订阅全服topic
		d.subscribe("service", ""),
		d.subscribe(broker.Path(d.ctx, "service", "balance"), "balance"),
		// 订阅服务类型topic
		d.subscribe(broker.Path(d.ctx, "service", d.ctx.GetName()), ""),
		d.subscribe(broker.Path(d.ctx, "service", "balance", d.ctx.GetName()), "balance"),
		// 订阅服务节点topic
		d.subscribe(d.GetAddress(), ""),
	)

	// 注册服务
	err = d.registry.Register(d.ctx, d.service, d.Options.RefreshInterval*2)
	if err != nil {
		log.Panicf(d.ctx, "register service %q node %q failed, %s", d.ctx.GetName(), d.ctx.GetId(), err)
	}
	log.Infof(d.ctx, "register service %q node %q success", d.ctx.GetName(), d.ctx.GetId())

	// 开始运行
	d.wg.Add(1)
	go d.run()
}

// ShutSP 关闭服务插件
func (d *_Distributed) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut service plugin %q", Name)

	d.wg.Wait()
}

// SendMsg 发送消息
func (d *_Distributed) SendMsg(dst string, msg gap.Msg) error {
	seq := atomic.AddInt64(&d.sendSeq, 1)

	mpBuf, err := codec.DefaultEncoder().EncodeBytes(seq, msg)
	if err != nil {
		return err
	}
	defer mpBuf.Release()

	return d.broker.Publish(context.Background(), dst, mpBuf.Data())
}

// GetFutures 获取异步模型Future控制器
func (d *_Distributed) GetFutures() concurrent.IFutures {
	return &d.futures
}

// GetAddress 获取服务节点地址
func (d *_Distributed) GetAddress() string {
	return d.service.Nodes[0].Address
}

func (d *_Distributed) subscribe(topic, queue string) broker.Subscriber {
	sub, err := d.broker.Subscribe(d.ctx, topic,
		broker.Option{}.EventHandler(generic.CastDelegateFunc1(d.handleEvent)),
		broker.Option{}.Queue(queue))
	if err != nil {
		log.Panicf(d.ctx, "subscribe topic %q queue %q failed, %s", topic, queue, err)
	}
	log.Infof(d.ctx, "subscribe topic %q queue %q success", topic, queue)
	return sub
}
