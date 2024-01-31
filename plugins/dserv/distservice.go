package dserv

import (
	"context"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/broker"
	"git.golaxy.org/framework/plugins/discovery"
	"git.golaxy.org/framework/plugins/dsync"
	"git.golaxy.org/framework/plugins/gap"
	"git.golaxy.org/framework/plugins/gap/codec"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/util/concurrent"
	"github.com/josharian/intern"
	"sync"
)

// Address 地址信息
type Address struct {
	Domain              string // 主域
	BroadcastSubdomain  string // 广播地址子域
	BalanceSubdomain    string // 负载均衡地址子域
	NodeSubdomain       string // 服务节点地址子域
	GlobalBroadcastAddr string // 全局广播地址
	GlobalBalanceAddr   string // 全局负载均衡地址
	BroadcastAddr       string // 服务广播地址
	BalanceAddr         string // 服务负载均衡地址
	LocalAddr           string // 本服务节点地址
}

// IWatcher 监听器
type IWatcher interface {
	context.Context
	Stop() <-chan struct{}
}

// IDistService 分布式服务支持
type IDistService interface {
	// GetAddress 获取地址信息
	GetAddress() Address
	// GetFutures 获取异步模型Future控制器
	GetFutures() concurrent.IFutures
	// MakeBroadcastAddr 创建服务广播地址
	MakeBroadcastAddr(service string) string
	// MakeBalanceAddr 创建服务负载均衡地址
	MakeBalanceAddr(service string) string
	// MakeNodeAddr 创建服务节点地址
	MakeNodeAddr(node string) (string, error)
	// SendMsg 发送消息
	SendMsg(dst string, msg gap.Msg) error
	// WatchMsg 监听消息（优先级高）
	WatchMsg(ctx context.Context, handler RecvMsgHandler) IWatcher
}

func newDistService(setting ...option.Setting[DistServiceOptions]) IDistService {
	return &_DistService{
		options: option.Make(Option{}.Default(), setting...),
	}
}

type _DistService struct {
	ctx           context.Context
	cancel        context.CancelFunc
	servCtx       service.Context
	wg            sync.WaitGroup
	options       DistServiceOptions
	registry      discovery.IRegistry
	broker        broker.IBroker
	dsync         dsync.IDistSync
	address       Address
	encoder       codec.Encoder
	decoder       codec.Decoder
	futures       concurrent.Futures
	sendMutex     sync.Mutex
	deduplication concurrent.Deduplication
	msgWatchers   concurrent.LockedSlice[*_MsgWatcher]
}

// InitSP 初始化服务插件
func (d *_DistService) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", Name)

	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.servCtx = ctx

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
	d.address = Address{Domain: d.options.Domain}
	d.address.BroadcastSubdomain = intern.String(broker.Path(d.servCtx, d.address.Domain, "broadcast"))
	d.address.BalanceSubdomain = intern.String(broker.Path(d.servCtx, d.address.Domain, "balance"))
	d.address.NodeSubdomain = intern.String(broker.Path(d.servCtx, d.address.Domain, "node"))
	d.address.GlobalBroadcastAddr = d.address.BroadcastSubdomain
	d.address.GlobalBalanceAddr = d.address.BalanceSubdomain
	d.address.BroadcastAddr = d.MakeBroadcastAddr(d.servCtx.GetName())
	d.address.BalanceAddr = d.MakeBalanceAddr(d.servCtx.GetName())
	d.address.LocalAddr, _ = d.MakeNodeAddr(d.servCtx.GetId().String())

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
	if !errors.Is(err, discovery.ErrNotFound) {
		log.Panicf(d.servCtx, "check service %q node %q failed, %s", d.servCtx.GetName(), d.servCtx.GetId(), err)
	}

	// 订阅topic
	subs := []broker.ISubscriber{
		// 订阅全服topic
		d.subscribe(d.address.GlobalBroadcastAddr, ""),
		d.subscribe(d.address.GlobalBalanceAddr, "balance"),
		// 订阅服务类型topic
		d.subscribe(d.address.BroadcastAddr, ""),
		d.subscribe(d.address.BalanceAddr, "balance"),
		// 订阅服务节点topic
		d.subscribe(d.address.LocalAddr, ""),
	}

	// 服务节点信息
	serviceNode := &discovery.Service{
		Name: d.servCtx.GetName(),
		Nodes: []discovery.Node{
			{
				Id:      d.servCtx.GetId().String(),
				Address: d.address.LocalAddr,
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
	log.Infof(ctx, "shut plugin %q", plugin.Name)

	d.cancel()
	d.wg.Wait()
}

// GetAddress 获取地址信息
func (d *_DistService) GetAddress() Address {
	return d.address
}

// GetFutures 获取异步模型Future控制器
func (d *_DistService) GetFutures() concurrent.IFutures {
	return &d.futures
}

// MakeBroadcastAddr 创建服务广播地址
func (d *_DistService) MakeBroadcastAddr(service string) string {
	return intern.String(broker.Path(d.servCtx, d.address.BroadcastSubdomain, service))
}

// MakeBalanceAddr 创建服务负载均衡地址
func (d *_DistService) MakeBalanceAddr(service string) string {
	return intern.String(broker.Path(d.servCtx, d.address.BalanceSubdomain, service))
}

// MakeNodeAddr 创建服务节点地址
func (d *_DistService) MakeNodeAddr(node string) (string, error) {
	if node == "" {
		return "", fmt.Errorf("%w: node is empty", core.ErrArgs)
	}
	return intern.String(broker.Path(d.servCtx, d.address.NodeSubdomain, node)), nil
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

		seq = d.deduplication.MakeSeq()
	}

	mpBuf, err := d.encoder.EncodeBytes(d.address.LocalAddr, seq, msg)
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
		broker.Option{}.EventHandler(generic.CastDelegateFunc1(d.handleEvent)),
		broker.Option{}.Queue(queue))
	if err != nil {
		log.Panicf(d.servCtx, "subscribe topic %q queue %q failed, %s", topic, queue, err)
	}
	log.Debugf(d.servCtx, "subscribe topic %q queue %q success", topic, queue)
	return sub
}
