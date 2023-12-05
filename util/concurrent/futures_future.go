package concurrent

import (
	"golang.org/x/net/context"
	"time"
)

// MakeFuture 创建Future
func MakeFuture[T Resp](fs IFutures, ctx context.Context, resp T, timeout ...time.Duration) Future {
	if ctx == nil {
		ctx = context.Background()
	}

	_timeout := fs.ptr().Timeout
	if len(timeout) > 0 {
		_timeout = timeout[0]
	}

	task := newTask(fs.ptr(), resp)
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
	f.futures.Resolve(f.Id, Ret[any]{Error: err})
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
