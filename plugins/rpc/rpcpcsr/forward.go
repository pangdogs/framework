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
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/plugins/rpcstack"
)

// PermissionValidator 权限验证器
type PermissionValidator = generic.DelegateFunc2[rpcstack.CallChain, callpath.CallPath, bool]

// NewForwardProcessor RPC转发处理器，用于S<->G的通信
func NewForwardProcessor(transitService string, mc gap.IMsgCreator, permValidator PermissionValidator) any {
	return &_ForwardProcessor{
		encoder:        codec.MakeEncoder(),
		decoder:        codec.MakeDecoder(mc),
		transitService: transitService,
		permValidator:  permValidator,
	}
}

// _ForwardProcessor RPC转发处理器，用于S<->G的通信
type _ForwardProcessor struct {
	servCtx              service.Context
	dist                 dserv.IDistService
	dentq                dentq.IDistEntityQuerier
	encoder              codec.Encoder
	decoder              codec.Decoder
	transitService       string
	transitBroadcastAddr string
	permValidator        PermissionValidator
	watcher              dserv.IWatcher
}

// Init 初始化
func (p *_ForwardProcessor) Init(ctx service.Context) {
	p.servCtx = ctx
	p.dist = dserv.Using(ctx)
	p.dentq = dentq.Using(ctx)
	p.transitBroadcastAddr = p.dist.MakeBroadcastAddr(p.transitService)
	p.watcher = p.dist.WatchMsg(context.Background(), generic.MakeDelegateFunc2(p.handleMsg))

	log.Debugf(p.servCtx, "rpc processor %q started", types.FullName(*p))
}

// Shut 结束
func (p *_ForwardProcessor) Shut(ctx service.Context) {
	<-p.watcher.Terminate()

	log.Debugf(p.servCtx, "rpc processor %q stopped", types.FullName(*p))
}
