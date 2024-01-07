package concurrent

import (
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy/runtime"
	"time"
)

// MakeRespAsyncRet 创建接收响应返回值的异步调用结果
func MakeRespAsyncRet() RespAsyncRet {
	return make(chan runtime.Ret, 1)
}

// MakeFutureRespAsyncRet 创建future与接收响应返回值的异步调用结果
func MakeFutureRespAsyncRet(fs IFutures, ctx context.Context, timeout ...time.Duration) (Future, RespAsyncRet) {
	resp := MakeRespAsyncRet()
	future := MakeFuture(fs, ctx, resp, timeout...)
	return future, resp
}

// RespAsyncRet 接收响应返回值的异步调用结果
type RespAsyncRet chan runtime.Ret

// Push 填入返回结果
func (resp RespAsyncRet) Push(ret Ret[any]) error {
	resp <- runtime.MakeRet(ret.Value, ret.Error)
	close(resp)
	return nil
}

// CastAsyncRet 转换为异步调用结果
func (resp RespAsyncRet) CastAsyncRet() runtime.AsyncRet {
	return chan runtime.Ret(resp)
}
