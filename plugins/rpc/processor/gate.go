package processor

import (
	"context"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"git.golaxy.org/framework/plugins/dentq"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/router"
	"git.golaxy.org/framework/util/concurrent"
)

// NewGateProcessor 创建网关RPC处理器，用于C<->G的通信
func NewGateProcessor(mc gap.IMsgCreator) any {
	return &_GateProcessor{
		encoder: codec.MakeEncoder(),
		decoder: codec.MakeDecoder(mc),
	}
}

// _GateProcessor 网关RPC处理器，用于C<->G的通信
type _GateProcessor struct {
	servCtx        service.Context
	dist           dserv.IDistService
	dentq          dentq.IDistEntityQuerier
	gate           gate.IGate
	router         router.IRouter
	encoder        codec.Encoder
	decoder        codec.Decoder
	sessionWatcher gate.IWatcher
	msgWatcher     dserv.IWatcher
}

// Init 初始化
func (p *_GateProcessor) Init(ctx service.Context) {
	p.servCtx = ctx
	p.dist = dserv.Using(ctx)
	p.dentq = dentq.Using(ctx)
	p.gate = gate.Using(ctx)
	p.router = router.Using(ctx)
	p.sessionWatcher = p.gate.Watch(context.Background(), generic.MakeDelegateAction3(p.handleSessionChanged))
	p.msgWatcher = p.dist.WatchMsg(context.Background(), generic.MakeDelegateFunc2(p.handleMsg))

	log.Debugf(p.servCtx, "rpc processor %q started", types.AnyFullName(*p))
}

// Shut 结束
func (p *_GateProcessor) Shut(ctx service.Context) {
	<-p.sessionWatcher.Terminate()
	<-p.msgWatcher.Terminate()

	log.Debugf(p.servCtx, "rpc processor %q stopped", types.AnyFullName(*p))
}

// Match 是否匹配
func (p *_GateProcessor) Match(ctx service.Context, dst, path string, oneWay bool) bool {
	return false
}

// Request 请求
func (p *_GateProcessor) Request(ctx service.Context, dst, path string, args []any) runtime.AsyncRet {
	ret := concurrent.MakeRespAsyncRet()
	ret.Push(concurrent.MakeRet[any](nil, ErrUndeliverable))
	return ret.CastAsyncRet()
}

// Notify 通知
func (p *_GateProcessor) Notify(ctx service.Context, dst, path string, args []any) error {
	return ErrUndeliverable
}
