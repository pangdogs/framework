package concurrent

import (
	"context"
	"errors"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrFuturesClosed           = errors.New("futures already closed")                   // Future控制器已关闭
	ErrFutureNotFound          = errors.New("future not found")                         // Future未找到
	ErrFutureCanceled          = errors.New("future canceled")                          // Future被取消
	ErrFutureTimeout           = errors.New("future timeout")                           // Future超时
	ErrFutureRespIncorrectType = errors.New("future response has incorrect value type") // Future响应的返回值类型错误
	ErrFutureReplyClosed       = errors.New("future reply closed")                      // Future答复已关闭
)

type (
	RequestHandler = generic.Action1[Future] // Future请求处理器
)

// IFutures Future控制器接口
type IFutures interface {
	iFutures

	// Make 创建Future
	Make(ctx context.Context, resp Resp, timeout ...time.Duration) Future
	// Request 请求
	Request(ctx context.Context, handler RequestHandler, timeout ...time.Duration) async.AsyncRet
	// Resolve 解决
	Resolve(id int64, ret async.Ret) error
}

type iFutures interface {
	ptr() *Futures
}

// MakeFutures 创建Future控制器
func MakeFutures(ctx context.Context, timeout time.Duration) Futures {
	if ctx == nil {
		ctx = context.Background()
	}
	return Futures{
		Ctx:     ctx,
		Id:      rand.Int63(),
		Timeout: timeout,
	}
}

// Futures Future控制器
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

// Request 请求
func (fs *Futures) Request(ctx context.Context, handler RequestHandler, timeout ...time.Duration) async.AsyncRet {
	if ctx == nil {
		ctx = context.Background()
	}

	future, resp := MakeFutureRespAsyncRet(fs, ctx, timeout...)
	handler.Exec(future)

	return resp.CastAsyncRet()
}

// Resolve 解决
func (fs *Futures) Resolve(id int64, ret async.Ret) error {
	v, ok := fs.tasks.LoadAndDelete(id)
	if !ok {
		return ErrFutureNotFound
	}
	return v.(iTask).Resolve(ret)
}

func (fs *Futures) ptr() *Futures {
	return fs
}

func (fs *Futures) makeId() int64 {
	id := atomic.AddInt64(&fs.Id, 1)
	if id == 0 {
		id = atomic.AddInt64(&fs.Id, 1)
	}
	return id
}
