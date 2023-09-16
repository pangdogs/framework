package transport

import "golang.org/x/net/context"

// Future 异步模型Future
type Future struct {
	Ctx     context.Context // 上下文
	Id      int64           // Id
	futures *Futures
}

// Cancel 取消
func (f Future) Cancel(err error) {
	f.futures.Dispatching(f.Id, nil, err)
}

// Wait 等待
func (f Future) Wait() {
	<-f.Ctx.Done()
}
