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
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework/addins/dent"
	"git.golaxy.org/framework/addins/dsvc"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/router"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"go.uber.org/zap"
)

// NewGateProcessor 创建网关RPC处理器，用于C<->G的通信
func NewGateProcessor(mc gap.IMsgCreator) any {
	return &_GateProcessor{
		encoder: codec.NewEncoder(),
		decoder: codec.NewDecoder(mc),
	}
}

// _GateProcessor 网关RPC处理器，用于C<->G的通信
type _GateProcessor struct {
	svcCtx      service.Context
	dsvc        dsvc.IDistService
	dentq       dent.IDistEntityQuerier
	gate        gate.IGate
	router      router.IRouter
	encoder     *codec.Encoder
	decoder     *codec.Decoder
	stopping    async.FutureVoid
	stoppingCtx context.Context
	stopped     [2]async.Future
}

// Init 初始化
func (p *_GateProcessor) Init(svcCtx service.Context) {
	p.svcCtx = svcCtx
	p.dsvc = dsvc.AddIn.Require(svcCtx)
	p.dentq = dent.QuerierAddIn.Require(svcCtx)
	p.gate = gate.AddIn.Require(svcCtx)
	p.router = router.AddIn.Require(svcCtx)
	p.stopping = async.NewFutureVoid()
	p.stoppingCtx = p.stopping.Out().Context(context.Background())

	var err error
	p.stopped[0], err = p.gate.Watch(p.stoppingCtx, generic.CastDelegateVoid1(p.handleSessionEstablished))
	if err != nil {
		log.L(svcCtx).Panic("watch gate session failed",
			zap.Error(err),
			zap.String("processor", types.FullName(*p)))
	}

	p.stopped[1], err = p.dsvc.Listen(p.stoppingCtx, generic.CastDelegateVoid2(p.handleServiceMsg))
	if err != nil {
		log.L(svcCtx).Panic("listen rpc message failed",
			zap.Error(err),
			zap.String("processor", types.FullName(*p)))
	}

	log.L(p.svcCtx).Debug("rpc processor started", zap.String("processor", types.FullName(*p)))
}

// Shut 结束
func (p *_GateProcessor) Shut(svcCtx service.Context) {
	async.ReturnVoid(p.stopping)

	for _, f := range p.stopped {
		<-f.Done()
	}

	log.L(p.svcCtx).Debug("rpc processor stopped", zap.String("processor", types.FullName(*p)))
}
