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
	"errors"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrFuturesClosed           = errors.New("futures already closed")                   // Future控制器已关闭
	ErrFutureNotFound          = errors.New("future not found")                         // Future未找到
	ErrFutureCanceled          = errors.New("future canceled")                          // Future被取消
	ErrFutureTimeout           = errors.New("future timeout")                           // Future超时
	ErrFutureRespIncorrectType = errors.New("future response has incorrect value type") // Future响应的返回值类型错误
)

type (
	RequestHandler = generic.Action1[Future] // Future请求处理器
)

// NewFutures 创建Future控制器
func NewFutures(ctx context.Context, timeout time.Duration) *Futures {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Futures{
		ctx:     ctx,
		id:      rand.Int63(),
		timeout: timeout,
	}
}

// Futures Future控制器
type Futures struct {
	ctx     context.Context // 上下文
	id      int64           // 请求id生成器
	timeout time.Duration   // 请求超时时间
	tasks   sync.Map
}

// Make 创建Future
func (fs *Futures) Make(ctx context.Context, resp Resp, timeout ...time.Duration) Future {
	if ctx == nil {
		ctx = context.Background()
	}

	_timeout := fs.timeout
	if len(timeout) > 0 {
		_timeout = timeout[0]
	}

	task := newTask(fs, resp)
	go task.Run(ctx, _timeout)

	return task.Future()
}

// Request 请求
func (fs *Futures) Request(ctx context.Context, handler RequestHandler, timeout ...time.Duration) async.AsyncRet {
	if ctx == nil {
		ctx = context.Background()
	}

	future, resp := MakeFutureRespAsyncRet(fs, ctx, timeout...)
	handler.Exec(future)

	return resp.ToAsyncRet()
}

// Resolve 解决
func (fs *Futures) Resolve(id int64, ret async.Ret) error {
	v, ok := fs.tasks.LoadAndDelete(id)
	if !ok {
		return ErrFutureNotFound
	}
	return v.(iTask).Resolve(ret)
}

func (fs *Futures) makeId() int64 {
	id := atomic.AddInt64(&fs.id, 1)
	if id == 0 {
		id = atomic.AddInt64(&fs.id, 1)
	}
	return id
}
