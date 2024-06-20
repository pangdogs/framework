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
func MakeFutureRespAsyncRet(fs IFutures, ctx context.Context, timeout ...time.Duration) (Future, RespAsyncRet) {
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
