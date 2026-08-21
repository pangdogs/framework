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
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework/addins/dent"
	"git.golaxy.org/framework/addins/dsvc"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"go.uber.org/zap"
)

// NewForwardProcessor 创建服务与网关间的 RPC 转发处理器。
func NewForwardProcessor(transitService string, mc gap.IMsgCreator, permValidator PermissionValidator, reduceCallPath bool) any {
	return &_ForwardProcessor{
		encoder:        codec.NewEncoder(),
		decoder:        codec.NewDecoder(mc),
		transitService: transitService,
		permValidator:  permValidator,
		reduceCallPath: reduceCallPath,
	}
}

// _ForwardProcessor 在服务消息通道与中转服务之间转发 RPC。
type _ForwardProcessor struct {
	svcCtx               service.Context
	dsvc                 dsvc.IDistService
	dentq                dent.IDistEntityQuerier
	encoder              *codec.Encoder
	decoder              *codec.Decoder
	scope                *async.Scope
	stopped              async.Signal
	transitService       string
	transitBroadcastAddr string
	permValidator        PermissionValidator
	reduceCallPath       bool
}

// Init 获取依赖并启动分布式服务消息监听。
func (p *_ForwardProcessor) Init(svcCtx service.Context) {
	p.svcCtx = svcCtx
	p.dsvc = dsvc.AddIn.Require(svcCtx)
	p.dentq = dent.QuerierAddIn.Require(svcCtx)
	p.scope = async.NewScope(nil)
	p.transitBroadcastAddr = p.dsvc.NodeDetails().MakeBroadcastAddr(p.transitService)

	var err error
	p.stopped, err = p.dsvc.Listen(p.scope.Context(), generic.CastDelegateVoid2(p.handleServiceMsg))
	if err != nil {
		p.scope.Close()
		log.L(svcCtx).Panic("listen rpc message failed", zap.Error(err), zap.String("processor", types.FullName(*p)))
	}

	log.L(p.svcCtx).Debug("rpc processor started", zap.String("processor", types.FullName(*p)))
}

// Shut 停止监听并等待消息处理循环及已接收任务退出。
func (p *_ForwardProcessor) Shut(svcCtx service.Context) {
	p.scope.Close()

	<-p.stopped.Done()
	<-p.scope.Completion().Done()

	log.L(p.svcCtx).Debug("rpc processor stopped", zap.String("processor", types.FullName(*p)))
}
