package concurrent

import (
	"context"
	"git.golaxy.org/core/util/types"
	"time"
)

// MakeRespChan 创建接收响应返回值的channel
func MakeRespChan[T any]() RespChan[T] {
	return make(RespChan[T], 1)
}

// MakeFutureRespChan 创建future与接收响应返回值的channel
func MakeFutureRespChan[T any](fs IFutures, ctx context.Context, timeout ...time.Duration) (Future, RespChan[T]) {
	resp := MakeRespChan[T]()
	future := MakeFuture(fs, ctx, resp, timeout...)
	return future, resp
}

// RespChan 接收响应返回值的channel
type RespChan[T any] chan Ret[T]

// Push 填入返回结果
func (resp RespChan[T]) Push(ret Ret[any]) error {
	if !ret.OK() {
		resp <- MakeRet[T](types.ZeroT[T](), ret.Error)
		close(resp)
		return nil
	}

	v, ok := ret.Value.(T)
	if !ok {
		resp <- MakeRet[T](types.ZeroT[T](), ErrFutureRespIncorrectType)
		close(resp)
		return nil
	}

	resp <- MakeRet[T](v, nil)
	close(resp)
	return nil
}

// CastReply 转换为异步答复
func (resp RespChan[T]) CastReply() Reply[T] {
	return chan Ret[T](resp)
}

// Reply 异步答复
type Reply[T any] <-chan Ret[T]

// Wait 等待
func (reply Reply[T]) Wait(ctx context.Context) Ret[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case ret, ok := <-reply:
		if !ok {
			return MakeRet[T](types.ZeroT[T](), ErrFutureReplyClosed)
		}
		return ret
	case <-ctx.Done():
		return MakeRet[T](types.ZeroT[T](), context.Canceled)
	}
}
