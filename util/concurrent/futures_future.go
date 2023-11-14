package concurrent

import "golang.org/x/net/context"

// Future 异步模型Future
type Future struct {
	Finish  context.Context // 上下文
	Id      int64           // Id
	futures *Futures
}

// Cancel 取消
func (f Future) Cancel(err error) {
	f.futures.Dispatching(f.Id, Ret[any]{Error: err})
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
