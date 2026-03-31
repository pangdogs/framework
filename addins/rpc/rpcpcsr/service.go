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
	"git.golaxy.org/framework/addins/dsvc"
	"git.golaxy.org/framework/addins/log"
	"go.uber.org/zap"
)

// NewServiceProcessor 创建分布式服务间的RPC处理器
func NewServiceProcessor(permValidator PermissionValidator, reduceCallPath bool) any {
	return &_ServiceProcessor{
		permValidator:  permValidator,
		reduceCallPath: reduceCallPath,
	}
}

// _ServiceProcessor 分布式服务间的RPC处理器
type _ServiceProcessor struct {
	svcCtx         service.Context
	dsvc           dsvc.IDistService
	stopping       async.FutureVoid
	stopped        async.Future
	permValidator  PermissionValidator
	reduceCallPath bool
}

// Init 初始化
func (p *_ServiceProcessor) Init(svcCtx service.Context) {
	p.svcCtx = svcCtx
	p.dsvc = dsvc.AddIn.Require(svcCtx)
	p.stopping = async.NewFutureVoid()

	var err error
	p.stopped, err = p.dsvc.Listen(p.stopping.Out().Context(context.Background()), generic.CastDelegateVoid2(p.handleServiceMsg))
	if err != nil {
		log.L(svcCtx).Panic("listen rpc message failed", zap.Error(err), zap.String("processor", types.FullName(*p)))
	}

	log.L(p.svcCtx).Debug("rpc processor started", zap.String("processor", types.FullName(*p)))
}

// Shut 结束
func (p *_ServiceProcessor) Shut(svcCtx service.Context) {
	async.ReturnVoid(p.stopping)

	<-p.stopped.Done()

	log.L(p.svcCtx).Debug("rpc processor stopped", zap.String("processor", types.FullName(*p)))
}
