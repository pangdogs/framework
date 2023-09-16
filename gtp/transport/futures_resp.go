package transport

import (
	"kit.golaxy.org/golaxy/runtime"
	"kit.golaxy.org/golaxy/util"
)

// Resp 响应接口
type Resp interface {
	// Push 填入返回结果
	Push(v any, err error) error
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
			ret.Error = ErrFutureRespIncorrectType
		}
	}
	return
}

// RespChan 响应返回值的channel
type RespChan[T any] chan Ret[T]

// Push 填入返回结果
func (resp RespChan[T]) Push(rv any, err error) (retErr error) {
	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			retErr = panicErr
		}
	}()

	resp <- newRet[T](rv, err)
	close(resp)

	return nil
}

// RespHandler 接收响应返回值的处理器
type RespHandler[T any] func(ret Ret[T])

// Push 填入返回结果
func (resp RespHandler[T]) Push(rv any, err error) (retErr error) {
	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			retErr = panicErr
		}
	}()

	resp(newRet[T](rv, err))

	return nil
}

// RespAsyncRet 接收响应返回值的异步调用结果
type RespAsyncRet chan runtime.Ret

// Push 填入返回结果
func (resp RespAsyncRet) Push(rv any, err error) (retErr error) {
	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			retErr = panicErr
		}
	}()

	resp <- runtime.NewRet(rv, err)
	close(resp)

	return nil
}

// Cast 转换为异步调用结果
func (resp RespAsyncRet) Cast() runtime.AsyncRet {
	return chan runtime.Ret(resp)
}
