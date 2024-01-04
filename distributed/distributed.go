package distributed

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy"
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
	"sync"
)

// Address 地址信息
type Address struct {
	Domain               string // 主域
	BroadcastSubdomain   string // 广播地址子域
	BalanceSubdomain     string // 负载均衡地址子域
	NodeSubdomain        string // 服务节点地址子域
	GlobalBroadcastAddr  string // 全局广播地址
	GlobalBalanceAddr    string // 全局负载均衡地址
	ServiceBroadcastAddr string // 服务广播地址
	ServiceBalanceAddr   string // 服务负载均衡地址
	LocalAddr            string // 本服务节点地址
}

// Watcher 监听器
type Watcher interface {
	context.Context
	Stop() <-chan struct{}
}

// Distributed 分布式服务支持
type Distributed interface {
	// GetAddress 获取地址信息
	GetAddress() Address
	// GetFutures 获取异步模型Future控制器
	GetFutures() concurrent.IFutures
	// MakeServiceBroadcastAddr 创建服务广播地址
	MakeServiceBroadcastAddr(service string) string
	// MakeServiceBalanceAddr 创建服务负载均衡地址
	MakeServiceBalanceAddr(service string) string
	// MakeServiceNodeAddr 创建服务节点地址
	MakeServiceNodeAddr(service, node string) (string, error)
	// SendMsg 发送消息
	SendMsg(dst string, msg gap.Msg) error
	// WatchMsg 监听消息
	WatchMsg(ctx context.Context, handler RecvMsgHandler) Watcher
}

func newDistributed(setting ...option.Setting[DistributedOptions]) Distributed {
	return &_Distributed{
		options: option.Make(Option{}.Default(), setting...),
	}
}

type _Distributed struct {
	ctx           context.Context
	cancel        context.CancelFunc
	servCtx       service.Context
	wg            sync.WaitGroup
	options       DistributedOptions
	registry      registry.Registry
	broker        broker.Broker
	dsync         dsync.DSync
	address       Address
	encoder       codec.Encoder
	decoder       codec.Decoder
	futures       concurrent.Futures
	deduplication concurrent.Deduplication
	msgWatchers   concurrent.LockedSlice[*_MsgWatcher]
}

// InitSP 初始化服务插件
func (d *_Distributed) InitSP(ctx service.Context) {
	log.Infof(ctx, "init service plugin <%s>:[%s]", Name, types.AnyFullName(*d))

	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.servCtx = ctx

	// 获取依赖的插件
	d.registry = registry.Using(d.servCtx)
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
	d.address = Address{Domain: d.options.Domain}
	d.address.BroadcastSubdomain = broker.Path(d.servCtx, d.address.Domain, "broadcast")
	d.address.BalanceSubdomain = broker.Path(d.servCtx, d.address.Domain, "balance")
	d.address.NodeSubdomain = broker.Path(d.servCtx, d.address.Domain, "node")
	d.address.GlobalBroadcastAddr = d.address.BroadcastSubdomain
	d.address.GlobalBalanceAddr = d.address.BalanceSubdomain
	d.address.ServiceBroadcastAddr = d.MakeServiceBroadcastAddr(d.servCtx.GetName())
	d.address.ServiceBalanceAddr = d.MakeServiceBalanceAddr(d.servCtx.GetName())
	d.address.LocalAddr, _ = d.MakeServiceNodeAddr(d.servCtx.GetName(), d.servCtx.GetId().String())

	// 加分布式锁
	mutex := d.dsync.NewMutex(dsync.Path(d.servCtx, "service", d.servCtx.GetName(), d.servCtx.GetId().String()))
	if err := mutex.Lock(d.servCtx); err != nil {
		log.Panicf(d.servCtx, "lock dsync mutex %q failed, %s", mutex.Name(), err)
	}
	defer mutex.Unlock(context.Background())

	// 检查服务节点是否已被注册
	_, err := d.registry.GetServiceNode(d.servCtx, d.servCtx.GetName(), d.servCtx.GetId().String())
	if err == nil {
		log.Panicf(d.servCtx, "check service %q node %q failed, already registered", d.servCtx.GetName(), d.servCtx.GetId())
	}
	if !errors.Is(err, registry.ErrNotFound) {
		log.Panicf(d.servCtx, "check service %q node %q failed, %s", d.servCtx.GetName(), d.servCtx.GetId(), err)
	}

	// 订阅topic
	subs := []broker.Subscriber{
		// 订阅全服topic
		d.subscribe(d.address.GlobalBroadcastAddr, ""),
		d.subscribe(d.address.GlobalBalanceAddr, "balance"),
		// 订阅服务类型topic
		d.subscribe(d.address.ServiceBroadcastAddr, ""),
		d.subscribe(d.address.ServiceBalanceAddr, "balance"),
		// 订阅服务节点topic
		d.subscribe(d.address.LocalAddr, ""),
	}

	// 服务节点信息
	serviceNode := registry.Service{
		Name:      d.servCtx.GetName(),
		Version:   d.options.Version,
		Metadata:  d.options.Metadata,
		Endpoints: d.options.Endpoints,
		Nodes: []registry.Node{
			{
				Id:      d.servCtx.GetId().String(),
				Address: d.address.LocalAddr,
			},
		},
	}

	// 注册服务
	err = d.registry.Register(d.servCtx, serviceNode, d.options.RefreshInterval*2)
	if err != nil {
		log.Panicf(d.servCtx, "register service %q node %q failed, %s", d.servCtx.GetName(), d.servCtx.GetId(), err)
	}
	log.Debugf(d.servCtx, "register service %q node %q success", d.servCtx.GetName(), d.servCtx.GetId())

	// 监控服务节点变化
	watcher, err := d.registry.Watch(d.ctx, "")
	if err != nil {
		log.Panicf(d.servCtx, "watching service changes failed, %s", err)
	}

	// 运行服务节点监听线程
	d.wg.Add(1)
	go d.watchingService(watcher)

	// 运行主线程
	d.wg.Add(1)
	go d.mainLoop(serviceNode, subs)
}

// ShutSP 关闭服务插件
func (d *_Distributed) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut service plugin <%s>:[%s]", plugin.Name, types.AnyFullName(*d))

	d.cancel()
	d.wg.Wait()
}

// GetAddress 获取地址信息
func (d *_Distributed) GetAddress() Address {
	return d.address
}

// GetFutures 获取异步模型Future控制器
func (d *_Distributed) GetFutures() concurrent.IFutures {
	return &d.futures
}

// MakeServiceBroadcastAddr 创建服务广播地址
func (d *_Distributed) MakeServiceBroadcastAddr(service string) string {
	return broker.Path(d.servCtx, d.address.BroadcastSubdomain, service)
}

// MakeServiceBalanceAddr 创建服务负载均衡地址
func (d *_Distributed) MakeServiceBalanceAddr(service string) string {
	return broker.Path(d.servCtx, d.address.BalanceSubdomain, service)
}

// MakeServiceNodeAddr 创建服务节点地址
func (d *_Distributed) MakeServiceNodeAddr(service, node string) (string, error) {
	if node == "" {
		return "", fmt.Errorf("%w: node is empty", golaxy.ErrArgs)
	}
	return broker.Path(d.servCtx, d.address.NodeSubdomain, service, node), nil
}

// SendMsg 发送消息
func (d *_Distributed) SendMsg(dst string, msg gap.Msg) error {
	if msg == nil {
		return fmt.Errorf("%w: msg is nil", golaxy.ErrArgs)
	}

	mpBuf, err := d.encoder.EncodeBytes(d.address.LocalAddr, d.deduplication.MakeSeq(), msg)
	if err != nil {
		return err
	}
	defer mpBuf.Release()

	return d.broker.Publish(d.ctx, dst, mpBuf.Data())
}

// WatchMsg 监听消息
func (d *_Distributed) WatchMsg(ctx context.Context, handler RecvMsgHandler) Watcher {
	return d.newMsgWatcher(ctx, handler)
}

func (d *_Distributed) subscribe(topic, queue string) broker.Subscriber {
	sub, err := d.broker.Subscribe(d.ctx, topic,
		broker.Option{}.EventHandler(generic.CastDelegateFunc1(d.handleEvent)),
		broker.Option{}.Queue(queue))
	if err != nil {
		log.Panicf(d.servCtx, "subscribe topic %q queue %q failed, %s", topic, queue, err)
	}
	log.Debugf(d.servCtx, "subscribe topic %q queue %q success", topic, queue)
	return sub
}
