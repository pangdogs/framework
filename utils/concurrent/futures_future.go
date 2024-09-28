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

package concurrent

import (
	"context"
	"git.golaxy.org/core/utils/async"
	"github.com/elliotchance/pie/v2"
	"time"
)

// MakeFuture 创建Future
func MakeFuture[T Resp](fs *Futures, ctx context.Context, resp T, timeout ...time.Duration) Future {
	if ctx == nil {
		ctx = context.Background()
	}

	_timeout := pie.First(timeout)
	if _timeout <= 0 {
		_timeout = fs.timeout
	}

	task := newTask(fs, resp)
	go task.Run(ctx, _timeout)

	return task.Future()
}

// Future 异步模型Future
type Future struct {
	Finish  context.Context // 上下文
	Id      int64           // Id
	futures *Futures
}

// Cancel 取消
func (f Future) Cancel(err error) {
	f.futures.Resolve(f.Id, async.MakeRet(nil, err))
}

// Wait 等待
func (f Future) Wait(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-ctx.Done():
	case <-f.Finish.Done():
	}
}
