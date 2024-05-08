package processor

import (
	"context"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"git.golaxy.org/framework/plugins/dentq"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/rpc/callpath"
)

// PermissionValidator 权限验证器
type PermissionValidator = generic.DelegateFunc2[string, callpath.CallPath, bool]

// NewForwardProcessor RPC转发处理器，用于S<->G的通信
func NewForwardProcessor(deliverService string, mc gap.IMsgCreator, permValidator PermissionValidator) any {
	return &_ForwardProcessor{
		encoder:        codec.MakeEncoder(),
		decoder:        codec.MakeDecoder(mc),
		deliverService: deliverService,
		permValidator:  permValidator,
	}
}

// _ForwardProcessor RPC转发处理器，用于S<->G的通信
type _ForwardProcessor struct {
	servCtx         service.Context
	dist            dserv.IDistService
	dentq           dentq.IDistEntityQuerier
	encoder         codec.Encoder
	decoder         codec.Decoder
	deliverService  string
	multicastBCAddr string
	permValidator   PermissionValidator
	watcher         dserv.IWatcher
}

// Init 初始化
func (p *_ForwardProcessor) Init(ctx service.Context) {
	p.servCtx = ctx
	p.dist = dserv.Using(ctx)
	p.dentq = dentq.Using(ctx)
	p.multicastBCAddr = p.dist.MakeBroadcastAddr(p.deliverService)
	p.watcher = p.dist.WatchMsg(context.Background(), generic.MakeDelegateFunc2(p.handleMsg))

	log.Debugf(p.servCtx, "rpc processor %q started", types.AnyFullName(*p))
}

// Shut 结束
func (p *_ForwardProcessor) Shut(ctx service.Context) {
	<-p.watcher.Terminate()

	log.Debugf(p.servCtx, "rpc processor %q stopped", types.AnyFullName(*p))
}
