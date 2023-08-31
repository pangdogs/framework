package transport

import (
	"context"
	"errors"
	"kit.golaxy.org/golaxy/util/concurrent"
	"sync/atomic"
	"time"
)

var (
	ErrAsyncReqNotFound       = errors.New("async request not found")                 // 异步请求未找到
	ErrAsyncReqCancelled      = errors.New("async request cancelled")                 // 异步请求被取消
	ErrAsyncReqTimeout        = errors.New("async request timeout")                   // 异步请求超时没有响应
	ErrAsyncRespIncorrectType = errors.New("async response has incorrect value type") // 异步响应的返回值类型错误
)

// IAsyncResp 异步响应接口
type IAsyncResp interface {
	// Push 填入返回结果
	Push(v any, err error)
}

// Ret 返回结果
type Ret[T any] struct {
	Value T     // 返回值
	Error error // 返回错误
}

func newRet[T any](v any, err error) (ret Ret[T]) {
	if err != nil {
		ret.Error = err
	} else {
		rv, ok := v.(T)
		if ok {
			ret.Value = rv
		} else {
			ret.Error = ErrAsyncRespIncorrectType
		}
	}
	return
}

// AsyncRespChan 异步响应接收返回值的channel
type AsyncRespChan[T any] chan Ret[T]

// Push 填入返回结果
func (resp AsyncRespChan[T]) Push(v any, err error) {
	resp <- newRet[T](v, err)
}

// AsyncRespHandler 接收返回值的处理器
type AsyncRespHandler[T any] func(ret Ret[T])

// Push 填入返回结果
func (resp AsyncRespHandler[T]) Push(v any, err error) {
	(resp)(newRet[T](v, err))
}

type _AsyncReq struct {
	Resp   IAsyncResp
	Finish context.CancelFunc
}

// AsyncDispatcher 异步请求响应分发器
type AsyncDispatcher struct {
	ReqId    int64                            // 请求id生成器
	Timeout  time.Duration                    // 请求超时时间
	waitResp concurrent.Map[int64, _AsyncReq] // 等待异步响应
}

// MakeRequest 创建异步请求
func (d *AsyncDispatcher) MakeRequest(ctx context.Context, resp IAsyncResp) (int64, error) {
	if resp == nil {
		return 0, errors.New("resp is nil")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	fin, cancel := context.WithCancel(context.Background())

	reqId := atomic.AddInt64(&d.ReqId, 1)
	d.waitResp.Store(reqId, _AsyncReq{Resp: resp, Finish: cancel})

	go func() {
		timer := time.NewTimer(d.Timeout)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			d.Dispatching(reqId, nil, ErrAsyncReqCancelled)
		case <-timer.C:
			d.Dispatching(reqId, nil, ErrAsyncReqTimeout)
		case <-fin.Done():
			return
		}
	}()

	return reqId, nil
}

// Dispatching 分发异步响应返回值
func (d *AsyncDispatcher) Dispatching(reqId int64, v any, err error) error {
	req, ok := d.waitResp.LoadAndDelete(reqId)
	if !ok {
		return ErrAsyncReqNotFound
	}
	req.Resp.Push(v, err)
	req.Finish()
	return nil
}
