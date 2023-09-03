package transport

import (
	"context"
	"errors"
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

// IAsyncResp 异步响应接口
type IAsyncResp interface {
	// Push 填入返回结果
	Push(v any, err error)
}

// AsyncRespChan 异步响应接收返回值的channel
type AsyncRespChan[T any] chan Ret[T]

// Push 填入返回结果
func (resp AsyncRespChan[T]) Push(rv any, err error) {
	resp <- newRet[T](rv, err)
	close(resp)
}

// AsyncRespHandler 接收返回值的处理器
type AsyncRespHandler[T any] func(ret Ret[T])

// Push 填入返回结果
func (resp AsyncRespHandler[T]) Push(rv any, err error) {
	(resp)(newRet[T](rv, err))
}

type _AsyncReq struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	Id     int64
	Resp   IAsyncResp
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

func (req *_AsyncReq) Reply(rv any, err error) {
	req.Cancel()
	if req.Resp != nil {
		req.Resp.Push(rv, err)
	}
}

func newAsyncReq(reqId int64, resp IAsyncResp) *_AsyncReq {
	ctx, cancel := context.WithCancel(context.Background())
	return &_AsyncReq{
		Ctx:    ctx,
		Cancel: cancel,
		Id:     reqId,
		Resp:   resp,
	}
}

// AsyncDispatcher 异步请求响应分发器
type AsyncDispatcher struct {
	ReqId    int64         // 请求id生成器
	Timeout  time.Duration // 请求超时时间
	waitResp sync.Map      // 等待异步响应
}

// MakeRequest 创建异步请求
func (d *AsyncDispatcher) MakeRequest(ctx context.Context, resp IAsyncResp) int64 {
	if ctx == nil {
		ctx = context.Background()
	}

	req := newAsyncReq(atomic.AddInt64(&d.ReqId, 1), resp)
	d.waitResp.Store(req.Id, req)

	go req.Wait(d, ctx)

	return req.Id
}

// Dispatching 分发异步响应返回值
func (d *AsyncDispatcher) Dispatching(reqId int64, rv any, err error) error {
	v, ok := d.waitResp.LoadAndDelete(reqId)
	if !ok {
		return ErrAsyncReqNotFound
	}
	v.(*_AsyncReq).Reply(rv, err)
	return nil
}
