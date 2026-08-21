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
	"errors"

	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/addins/log"
	"go.uber.org/zap"
)

func spawnProcessorTask(svcCtx service.Context, scope *async.Scope, task func(context.Context)) {
	async.SpawnVoid(scope, task).OnComplete(func(ret async.Result) {
		if ret.Error == nil || errors.Is(ret.Error, async.ErrScopeClosed) {
			return
		}
		log.L(svcCtx).Error("rpc processor task failed", zap.Error(ret.Error))
	})
}
