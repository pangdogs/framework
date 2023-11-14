package concurrent

import (
	"context"
	"errors"
	"kit.golaxy.org/golaxy/runtime"
	"kit.golaxy.org/golaxy/util/generic"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrFuturesClosed           = errors.New("futures already closed")                   // 异步模型Future已关闭
	ErrFutureNotFound          = errors.New("future not found")                         // Future未找到
	ErrFutureCancelled         = errors.New("future cancelled")                         // Future被取消
	ErrFutureTimeout           = errors.New("future timeout")                           // Future超时
	ErrFutureRespIncorrectType = errors.New("future response has incorrect value type") // Future响应的返回值类型错误
)

type (
	RequestHandler = generic.Action1[Future] // Future请求处理器
)

// MakeFuture 创建异步模型Future
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

// IFutures 异步模型Future控制器接口
type IFutures interface {
	// Make 创建Future
	Make(ctx context.Context, resp Resp, timeout ...time.Duration) Future
	// MakeId 创建请求Id
	MakeId() int64
	// Request 异步请求
	Request(ctx context.Context, handler RequestHandler, timeout ...time.Duration) runtime.AsyncRet
	// Dispatching 分发异步响应返回值
	Dispatching(id int64, ret Ret[any]) error

	ptr() *Futures
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

	task := newTask(fs, resp)
	go task.Run(ctx, _timeout)

	return task.Future()
}

// MakeId 创建请求Id
func (fs *Futures) MakeId() int64 {
	return atomic.AddInt64(&fs.Id, 1)
}

// Request 异步请求
func (fs *Futures) Request(ctx context.Context, handler RequestHandler, timeout ...time.Duration) runtime.AsyncRet {
	if ctx == nil {
		ctx = context.Background()
	}

	asyncRet := make(RespAsyncRet, 1)
	handler.Exec(fs.Make(ctx, asyncRet, timeout...))

	return asyncRet.Cast()
}

// Dispatching 分发异步响应返回值
func (fs *Futures) Dispatching(id int64, ret Ret[any]) error {
	v, ok := fs.tasks.LoadAndDelete(id)
	if !ok {
		return ErrFutureNotFound
	}
	return v.(_ITask).Reply(ret)
}
