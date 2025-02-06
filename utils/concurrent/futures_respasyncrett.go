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
	"git.golaxy.org/core/utils/types"
	"time"
)

// MakeRespAsyncRetT 创建接收响应返回值的异步调用结果
func MakeRespAsyncRetT[T any]() RespAsyncRetT[T] {
	return async.MakeAsyncRetT[T]()
}

// MakeFutureRespAsyncRetT 创建future与接收响应返回值的异步调用结果
func MakeFutureRespAsyncRetT[T any](fs *Futures, ctx context.Context, timeout ...time.Duration) (Future, RespAsyncRetT[T]) {
	resp := MakeRespAsyncRetT[T]()
	future := MakeFuture(fs, ctx, resp, timeout...)
	return future, resp
}

// RespAsyncRetT 接收响应返回值的channel
type RespAsyncRetT[T any] chan async.RetT[T]

// Push 填入返回结果
func (resp RespAsyncRetT[T]) Push(ret async.Ret) error {
	retT, ok := async.AsRetT[T](ret)
	if !ok {
		async.ReturnT(resp, async.MakeRetT[T](types.ZeroT[T](), ErrFutureRespIncorrectType))
		return nil
	}

	async.ReturnT(resp, retT)
	return nil
}

// ToAsyncRetT 转换为异步调用结果
func (resp RespAsyncRetT[T]) ToAsyncRetT() async.AsyncRetT[T] {
	return chan async.RetT[T](resp)
}
