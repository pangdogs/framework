package transport

import (
	"context"
	"errors"
	"kit.golaxy.org/golaxy/runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrFuturesClosed           = errors.New("futures closed")                           // 异步模型Future已关闭
	ErrFutureNotFound          = errors.New("future not found")                         // Future未找到
	ErrFutureCancelled         = errors.New("future cancelled")                         // Future被取消
	ErrFutureTimeout           = errors.New("future timeout")                           // Future超时
	ErrFutureRespIncorrectType = errors.New("future response has incorrect value type") // Future响应的返回值类型错误
)

// IFutures 异步模型Future控制器接口
type IFutures interface {
	// Make 创建Future
	Make(ctx context.Context, resp Resp, timeout ...time.Duration) Future
	// Request 异步请求
	Request(ctx context.Context, handler func(future Future), timeout ...time.Duration) runtime.AsyncRet
	// Dispatching 分发异步响应返回值
	Dispatching(id int64, rv any, err error) error
}

type _FutureTask struct {
	Future Future
	Resp   Resp
	Cancel context.CancelFunc
}

func (ft *_FutureTask) Run(ctx context.Context, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ft.Future.futures.Ctx.Done():
		ft.Future.futures.Dispatching(ft.Future.Id, nil, ErrFuturesClosed)
	case <-ctx.Done():
		ft.Future.futures.Dispatching(ft.Future.Id, nil, ErrFutureCancelled)
	case <-timer.C:
		ft.Future.futures.Dispatching(ft.Future.Id, nil, ErrFutureTimeout)
	case <-ft.Future.Ctx.Done():
		return
	}
}

func (ft *_FutureTask) Reply(rv any, err error) error {
	ft.Cancel()

	if ft.Resp != nil {
		return ft.Resp.Push(rv, err)
	}

	return nil
}

// Futures 异步模型Future控制器
type Futures struct {
	Ctx     context.Context // 上下文
	Id      int64           // 请求id生成器
	Timeout time.Duration   // 请求超时时间
	tasks   sync.Map
}

// Make 创建Future
func (fs *Futures) Make(ctx context.Context, resp Resp, timeout ...time.Duration) Future {
	if ctx == nil {
		ctx = context.Background()
	}

	_timeout := fs.Timeout
	if len(timeout) > 0 {
		_timeout = timeout[0]
	}

	task := fs.newFutureTask(resp)
	go task.Run(ctx, _timeout)

	return task.Future
}

// Request 异步请求
func (fs *Futures) Request(ctx context.Context, handler func(future Future), timeout ...time.Duration) runtime.AsyncRet {
	if ctx == nil {
		ctx = context.Background()
	}

	if handler == nil {
		panic("handler is nil")
	}

	asyncRet := make(RespAsyncRet, 1)
	handler(fs.Make(ctx, asyncRet, timeout...))

	return asyncRet.Cast()
}

// Dispatching 分发异步响应返回值
func (fs *Futures) Dispatching(id int64, rv any, err error) error {
	v, ok := fs.tasks.LoadAndDelete(id)
	if !ok {
		return ErrFutureNotFound
	}
	return v.(*_FutureTask).Reply(rv, err)
}

func (fs *Futures) newFutureTask(resp Resp) *_FutureTask {
	ctx, cancel := context.WithCancel(context.Background())

	wait := &_FutureTask{
		Future: Future{
			Ctx:     ctx,
			Id:      atomic.AddInt64(&fs.Id, 1),
			futures: fs,
		},
		Resp:   resp,
		Cancel: cancel,
	}
	fs.tasks.Store(wait.Future.Id, wait)

	return wait
}
