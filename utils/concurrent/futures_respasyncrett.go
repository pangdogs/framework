package concurrent

import (
	"context"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/types"
	"time"
)

// MakeRespAsyncRetT 创建接收响应返回值的异步调用结果
func MakeRespAsyncRetT[T any]() RespAsyncRetT[T] {
	return make(RespAsyncRetT[T], 1)
}

// MakeFutureRespAsyncRetT 创建future与接收响应返回值的异步调用结果
func MakeFutureRespAsyncRetT[T any](fs IFutures, ctx context.Context, timeout ...time.Duration) (Future, RespAsyncRetT[T]) {
	resp := MakeRespAsyncRetT[T]()
	future := MakeFuture(fs, ctx, resp, timeout...)
	return future, resp
}

// RespAsyncRetT 接收响应返回值的channel
type RespAsyncRetT[T any] chan async.RetT[T]

// Push 填入返回结果
func (ch RespAsyncRetT[T]) Push(ret async.Ret) error {
	resp, ok := async.AsRetT[T](ret)
	if !ok {
		ch <- async.MakeRetT[T](types.ZeroT[T](), ErrFutureRespIncorrectType)
		close(ch)
		return nil
	}

	ch <- resp
	close(ch)
	return nil
}

// ToAsyncRetT 转换为异步调用结果
func (ch RespAsyncRetT[T]) ToAsyncRetT() async.AsyncRetT[T] {
	return chan async.RetT[T](ch)
}
