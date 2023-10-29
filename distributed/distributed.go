package distributed

import (
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/plugins/broker"
	"kit.golaxy.org/plugins/dsync"
	"kit.golaxy.org/plugins/gtp/transport"
	"kit.golaxy.org/plugins/log"
	"kit.golaxy.org/plugins/registry"
	"math/rand"
	"strings"
	"sync"
)

// Distributed 分布式服务支持
type Distributed interface {
	// GetFutures 获取异步模型Future控制器
	GetFutures() transport.IFutures
}

func newDistributed(options ...DistributedOption) Distributed {
	opts := DistributedOptions{}
	Option{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_Distributed{}
}

type _Distributed struct {
	ctx            service.Context
	Options        DistributedOptions
	wg             sync.WaitGroup
	service        registry.Service
	pluginRegistry registry.Registry
	pluginBroker   broker.Broker
	pluginDSync    dsync.DSync
	futures        transport.Futures
}

// InitSP 初始化服务插件
func (d *_Distributed) InitSP(ctx service.Context) {
	log.Infof(ctx, "init service plugin %q with %q", definePlugin.Name, types.AnyFullName(*d))

	d.ctx = ctx

	// 获取依赖的插件
	d.pluginRegistry = registry.Fetch(ctx)
	d.pluginBroker = broker.Fetch(ctx)
	d.pluginDSync = dsync.Fetch(ctx)

	// 初始化异步模型Future控制器
	d.futures.Ctx = d.ctx
	d.futures.Id = rand.Int63()
	d.futures.Timeout = d.Options.FutureTimeout

	// 初始化服务信息
	d.service = registry.Service{
		Name: d.ctx.GetName(),
		Nodes: []registry.Node{
			{
				Id:      d.ctx.GetId().String(),
				Address: strings.Join([]string{"service", d.ctx.GetName(), d.ctx.GetId().String()}, d.pluginBroker.Separator()),
			},
		},
	}

	//
	mutex := d.pluginDSync.NewMutex(strings.Join([]string{"service", d.service.Name, d.ctx.GetId().String()}, d.pluginDSync.Separator()))
	if err := mutex.Lock(d.ctx); err != nil {
		log.Panicf(d.ctx, "")
	}
	defer mutex.Unlock(context.Background())

	err := d.pluginRegistry.Register(d.ctx, d.service, d.Options.RefreshInterval*2)
	if err != nil {
		log.Warnf(d.ctx, "register distributed service %q failed, %s", d.ctx, err)
	}

	subService, err := d.pluginBroker.Subscribe(d.ctx, "service", broker.Option{}.EventHandler(d.EventHandler))
	if err != nil {
		log.Panicf(d.ctx, "subscribe topic %q failed, %s", d.service.Name, err)
	}
	defer subService.Unsubscribe()

	subServiceName, err := d.pluginBroker.Subscribe(d.ctx, strings.Join([]string{"service", d.service.Name}, d.pluginBroker.Separator()), broker.Option{}.EventHandler(d.EventHandler))
	if err != nil {
		log.Panicf(d.ctx, "subscribe topic %q failed, %s", d.service.Name, err)
	}
	defer subServiceName.Unsubscribe()

	subServiceNode, err := d.pluginBroker.Subscribe(d.ctx, d.service.Nodes[0].Address, broker.Option{}.EventHandler(d.EventHandler))
	if err != nil {
		log.Panicf(d.ctx, "subscribe topic %q failed, %s", d.service.Nodes[0], err)
	}
	defer subServiceNode.Unsubscribe()

	d.wg.Add(1)
	go d.Run()
}

// ShutSP 关闭服务插件
func (d *_Distributed) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut service plugin %q", definePlugin.Name)

	d.wg.Wait()
}

// GetFutures 获取异步模型Future控制器
func (d *_Distributed) GetFutures() transport.IFutures {
	return &d.futures
}
