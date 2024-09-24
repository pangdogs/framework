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

package rpcpcsr

import (
	"context"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"git.golaxy.org/framework/plugins/dentq"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/gate"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/router"
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
	svcCtx         service.Context
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
func (p *_GateProcessor) Init(svcCtx service.Context) {
	p.svcCtx = svcCtx
	p.dist = dserv.Using(svcCtx)
	p.dentq = dentq.Using(svcCtx)
	p.gate = gate.Using(svcCtx)
	p.router = router.Using(svcCtx)
	p.sessionWatcher = p.gate.Watch(context.Background(), generic.MakeDelegateAction3(p.handleSessionChanged))
	p.msgWatcher = p.dist.WatchMsg(context.Background(), generic.MakeDelegateFunc2(p.handleMsg))

	log.Debugf(p.svcCtx, "rpc processor %q started", types.FullName(*p))
}

// Shut 结束
func (p *_GateProcessor) Shut(svcCtx service.Context) {
	<-p.sessionWatcher.Terminate()
	<-p.msgWatcher.Terminate()

	log.Debugf(p.svcCtx, "rpc processor %q stopped", types.FullName(*p))
}
