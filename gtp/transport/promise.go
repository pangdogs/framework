package transport

import (
	"context"
	"errors"
	"kit.golaxy.org/golaxy/util"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrPromiseClosed          = errors.New("promise closed")                          // promise已关闭
	ErrAsyncReqNotFound       = errors.New("async request not found")                 // 异步请求未找到
	ErrAsyncReqCancelled      = errors.New("async request cancelled")                 // 异步请求被取消
	ErrAsyncReqTimeout        = errors.New("async request timeout")                   // 异步请求超时没有响应
	ErrAsyncRespIncorrectType = errors.New("async response has incorrect value type") // 异步响应的返回值类型错误
)

// AsyncReq 异步请求
type AsyncReq struct {
	Ctx     context.Context // 上下文
	Id      int64           // 请求Id
	promise *Promise
}

// Interrupt 中断
func (req AsyncReq) Interrupt(err error) {
	req.promise.Dispatching(req.Id, nil, err)
}

// Ret 返回结果
type Ret[T any] struct {
	Value T     // 返回值
	Error error // 返回错误
}

func newRet[T any](rv any, err error) (ret Ret[T]) {
	if err != nil {
		ret.Error = err
	} else {
		value, ok := rv.(T)
		if ok {
			ret.Value = value
		} else {
			ret.Error = ErrAsyncRespIncorrectType
		}
	}
	return
}

// AsyncResp 异步响应接口
type AsyncResp interface {
	// Push 填入返回结果
	Push(v any, err error) error
}

// AsyncRespChan 异步响应接收返回值的channel
type AsyncRespChan[T any] chan Ret[T]

// Push 填入返回结果
func (resp AsyncRespChan[T]) Push(rv any, err error) (retErr error) {
	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			retErr = panicErr
		}
	}()

	resp <- newRet[T](rv, err)
	close(resp)

	return nil
}

// AsyncRespHandler 接收返回值的处理器
type AsyncRespHandler[T any] func(ret Ret[T])

// Push 填入返回结果
func (resp AsyncRespHandler[T]) Push(rv any, err error) (retErr error) {
	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			retErr = panicErr
		}
	}()

	resp(newRet[T](rv, err))

	return nil
}

type _AsyncWait struct {
	Req    AsyncReq
	Resp   AsyncResp
	Cancel context.CancelFunc
}

func (w *_AsyncWait) Run(ctx context.Context, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-w.Req.promise.Ctx.Done():
		w.Req.promise.Dispatching(w.Req.Id, nil, ErrPromiseClosed)
	case <-ctx.Done():
		w.Req.promise.Dispatching(w.Req.Id, nil, ErrAsyncReqCancelled)
	case <-timer.C:
		w.Req.promise.Dispatching(w.Req.Id, nil, ErrAsyncReqTimeout)
	case <-w.Req.Ctx.Done():
		return
	}
}

func (w *_AsyncWait) Reply(rv any, err error) error {
	w.Cancel()

	if w.Resp != nil {
		return w.Resp.Push(rv, err)
	}

	return nil
}

// IPromise 异步编程模型承诺（Promise）接口
type IPromise interface {
	// MakeRequest 创建异步请求
	MakeRequest(ctx context.Context, resp AsyncResp) AsyncReq
	// MakeRequestWithTimeout 使用自定义超时时间，创建异步请求
	MakeRequestWithTimeout(ctx context.Context, resp AsyncResp, timeout time.Duration) AsyncReq
	// Dispatching 分发异步响应返回值
	Dispatching(reqId int64, rv any, err error) error
}

// Promise 异步编程模型承诺（Promise）
type Promise struct {
	Ctx     context.Context // 上下文
	Id      int64           // 请求id生成器
	Timeout time.Duration   // 请求超时时间
	waits   sync.Map
}

// MakeRequest 创建异步请求
func (d *Promise) MakeRequest(ctx context.Context, resp AsyncResp) AsyncReq {
	if ctx == nil {
		ctx = context.Background()
	}

	wait := d.newWait(resp)
	d.waits.Store(wait.Req.Id, wait)

	go wait.Run(ctx, d.Timeout)

	return wait.Req
}

// MakeRequestWithTimeout 使用自定义超时时间，创建异步请求
func (d *Promise) MakeRequestWithTimeout(ctx context.Context, resp AsyncResp, timeout time.Duration) AsyncReq {
	if ctx == nil {
		ctx = context.Background()
	}

	wait := d.newWait(resp)
	d.waits.Store(wait.Req.Id, wait)

	go wait.Run(ctx, timeout)

	return wait.Req
}

// Dispatching 分发异步响应返回值
func (d *Promise) Dispatching(reqId int64, rv any, err error) error {
	v, ok := d.waits.LoadAndDelete(reqId)
	if !ok {
		return ErrAsyncReqNotFound
	}
	return v.(*_AsyncWait).Reply(rv, err)
}

func (d *Promise) newWait(resp AsyncResp) *_AsyncWait {
	ctx, cancel := context.WithCancel(context.Background())
	return &_AsyncWait{
		Req: AsyncReq{
			Ctx:     ctx,
			Id:      atomic.AddInt64(&d.Id, 1),
			promise: d,
		},
		Resp:   resp,
		Cancel: cancel,
	}
}
