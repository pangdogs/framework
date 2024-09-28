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
	"git.golaxy.org/framework/plugins/dsvc"
	"git.golaxy.org/framework/plugins/log"
)

// NewServiceProcessor 创建分布式服务间的RPC处理器
func NewServiceProcessor(permValidator PermissionValidator, reduceCP bool) any {
	return &_ServiceProcessor{
		permValidator: permValidator,
		reduceCP:      reduceCP,
	}
}

// _ServiceProcessor 分布式服务间的RPC处理器
type _ServiceProcessor struct {
	svcCtx        service.Context
	dist          dsvc.IDistService
	watcher       dsvc.IWatcher
	permValidator PermissionValidator
	reduceCP      bool
}

// Init 初始化
func (p *_ServiceProcessor) Init(svcCtx service.Context) {
	p.svcCtx = svcCtx
	p.dist = dsvc.Using(svcCtx)
	p.watcher = p.dist.WatchMsg(context.Background(), generic.MakeDelegateFunc2(p.handleMsg))

	log.Debugf(p.svcCtx, "rpc processor %q started", types.FullName(*p))
}

// Shut 结束
func (p *_ServiceProcessor) Shut(svcCtx service.Context) {
	<-p.watcher.Terminate()

	log.Debugf(p.svcCtx, "rpc processor %q stopped", types.FullName(*p))
}
