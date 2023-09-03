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
	ErrAsyncReqNotFound       = errors.New("async request not found")                 // 异步请求未找到
	ErrAsyncReqCancelled      = errors.New("async request cancelled")                 // 异步请求被取消
	ErrAsyncReqTimeout        = errors.New("async request timeout")                   // 异步请求超时没有响应
	ErrAsyncRespIncorrectType = errors.New("async response has incorrect value type") // 异步响应的返回值类型错误
)

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

type _AsyncReq struct {
	Ctx    context.Context
	cancel context.CancelFunc
	Id     int64
	Resp   AsyncResp
}

func (req *_AsyncReq) Wait(d *AsyncDispatcher, ctx context.Context) {
	timer := time.NewTimer(d.Timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		d.Dispatching(req.Id, nil, ErrAsyncReqCancelled)
	case <-timer.C:
		d.Dispatching(req.Id, nil, ErrAsyncReqTimeout)
	case <-req.Ctx.Done():
		return
	}
}

func (req *_AsyncReq) Reply(rv any, err error) error {
	req.cancel()

	if req.Resp != nil {
		return req.Resp.Push(rv, err)
	}

	return nil
}

func newAsyncReq(reqId int64, resp AsyncResp) *_AsyncReq {
	ctx, cancel := context.WithCancel(context.Background())
	return &_AsyncReq{
		Ctx:    ctx,
		cancel: cancel,
		Id:     reqId,
		Resp:   resp,
	}
}

// AsyncReq 异步请求
type AsyncReq struct {
	Ctx             context.Context
	Id              int64
	asyncDispatcher *AsyncDispatcher
}

// Interrupt 中断
func (req AsyncReq) Interrupt(err error) {
	req.asyncDispatcher.Dispatching(req.Id, nil, err)
}

// IAsyncDispatcher 异步请求响应分发器接口
type IAsyncDispatcher interface {
	// MakeRequest 创建异步请求
	MakeRequest(ctx context.Context, resp AsyncResp) AsyncReq
	// Dispatching 分发异步响应返回值
	Dispatching(reqId int64, rv any, err error) error
}

// AsyncDispatcher 异步请求响应分发器
type AsyncDispatcher struct {
	ReqId    int64         // 请求id生成器
	Timeout  time.Duration // 请求超时时间
	waitResp sync.Map      // 等待异步响应
}

// MakeRequest 创建异步请求
func (d *AsyncDispatcher) MakeRequest(ctx context.Context, resp AsyncResp) AsyncReq {
	if ctx == nil {
		ctx = context.Background()
	}

	req := newAsyncReq(atomic.AddInt64(&d.ReqId, 1), resp)
	d.waitResp.Store(req.Id, req)

	go req.Wait(d, ctx)

	return AsyncReq{
		Ctx: req.Ctx,
		Id:  req.Id,
	}
}

// Dispatching 分发异步响应返回值
func (d *AsyncDispatcher) Dispatching(reqId int64, rv any, err error) error {
	v, ok := d.waitResp.LoadAndDelete(reqId)
	if !ok {
		return ErrAsyncReqNotFound
	}
	return v.(*_AsyncReq).Reply(rv, err)
}
