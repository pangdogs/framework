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
	"time"
)

// MakeRespAsyncRet 创建接收响应返回值的异步调用结果
func MakeRespAsyncRet() RespAsyncRet {
	return make(chan async.Ret, 1)
}

// MakeFutureRespAsyncRet 创建future与接收响应返回值的异步调用结果
func MakeFutureRespAsyncRet(fs *Futures, ctx context.Context, timeout ...time.Duration) (Future, RespAsyncRet) {
	resp := MakeRespAsyncRet()
	future := MakeFuture(fs, ctx, resp, timeout...)
	return future, resp
}

// RespAsyncRet 接收响应返回值的异步调用结果
type RespAsyncRet chan async.Ret

// Push 填入返回结果
func (ch RespAsyncRet) Push(ret async.Ret) error {
	ch <- async.MakeRet(ret.Value, ret.Error)
	close(ch)
	return nil
}

// ToAsyncRet 转换为异步调用结果
func (ch RespAsyncRet) ToAsyncRet() async.AsyncRet {
	return chan async.Ret(ch)
}
