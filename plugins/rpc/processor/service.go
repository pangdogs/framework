package processor

import (
	"context"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/log"
)

// NewServiceProcessor 创建分布式服务间的RPC处理器
func NewServiceProcessor(permValidator PermissionValidator) any {
	return &_ServiceProcessor{
		permValidator: permValidator,
	}
}

// _ServiceProcessor 分布式服务间的RPC处理器
type _ServiceProcessor struct {
	servCtx       service.Context
	dist          dserv.IDistService
	watcher       dserv.IWatcher
	permValidator PermissionValidator
}

// Init 初始化
func (p *_ServiceProcessor) Init(ctx service.Context) {
	p.servCtx = ctx
	p.dist = dserv.Using(ctx)
	p.watcher = p.dist.WatchMsg(context.Background(), generic.MakeDelegateFunc2(p.handleMsg))

	log.Debugf(p.servCtx, "rpc processor %q started", types.FullName(*p))
}

// Shut 结束
func (p *_ServiceProcessor) Shut(ctx service.Context) {
	<-p.watcher.Terminate()

	log.Debugf(p.servCtx, "rpc processor %q stopped", types.FullName(*p))
}
