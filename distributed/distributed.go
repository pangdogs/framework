package distributed

import (
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/broker"
	"kit.golaxy.org/plugins/dsync"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/registry"
	"strings"
	"sync"
	"time"
)

// Distributed 分布式服务支持
type Distributed interface {
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
	pluginRegistry registry.Registry
	pluginBroker   broker.Broker
	pluginDSync    dsync.DSync
}

// InitSP 初始化服务插件
func (d *_Distributed) InitSP(ctx service.Context) {
	logger.Infof(ctx, "init service plugin %q with %q", definePlugin.Name, util.TypeOfAnyFullName(*d))

	d.ctx = ctx

	d.pluginRegistry = registry.Fetch(ctx)
	d.pluginBroker = broker.Fetch(ctx)
	d.pluginDSync = dsync.Fetch(ctx)

	d.wg.Add(1)
	go d.Run()
}

// ShutSP 关闭服务插件
func (d *_Distributed) ShutSP(ctx service.Context) {
	logger.Infof(ctx, "shut service plugin %q", definePlugin.Name)

	d.wg.Wait()
}

// Run 运行
func (d *_Distributed) Run() {
	ticker := time.NewTicker(d.Options.RefreshInterval)

	defer func() {
		ticker.Stop()
		d.wg.Done()
	}()

	service := registry.Service{
		Name: d.ctx.GetName(),
		Nodes: []registry.Node{
			{
				Id:      d.ctx.GetId().String(),
				Address: strings.Join([]string{d.ctx.GetName(), d.ctx.GetId().String()}, d.pluginBroker.GetSeparator()),
			},
		},
	}

loop:
	for {
		select {
		case <-ticker.C:
			err := d.pluginRegistry.Register(d.ctx, service, d.Options.RefreshInterval*2)
			if err != nil {
				logger.Warnf(d.ctx, "register service failed, %s", err)
			}
		case <-d.ctx.Done():
			break loop
		}
	}

	err := d.pluginRegistry.Deregister(context.Background(), service)
	if err != nil {
		logger.Warnf(d.ctx, "deregister service failed, %s", err)
	}
}
