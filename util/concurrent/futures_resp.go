package concurrent

import (
	"kit.golaxy.org/golaxy/runtime"
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/golaxy/util/types"
)

// Resp 响应接口
type Resp interface {
	// Push 填入返回结果
	Push(ret Ret[any]) error
}

// Ret 返回结果
type Ret[T any] struct {
	Value T     // 返回值
	Error error // 返回错误
}

func MakeRet[T any](v any, err error) (ret Ret[T]) {
	if err != nil {
		ret.Error = err
	} else {
		value, ok := v.(T)
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
func (resp RespChan[T]) Push(ret Ret[any]) (retErr error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			retErr = panicErr
		}
	}()

	resp <- MakeRet[T](ret.Value, ret.Error)
	close(resp)

	return nil
}

// RespHandler 接收响应返回值的处理器
type RespHandler[T any] generic.Action1[Ret[T]]

// Push 填入返回结果
func (resp RespHandler[T]) Push(ret Ret[any]) (retErr error) {
	return generic.CastAction1(resp).Invoke(MakeRet[T](ret.Value, ret.Error))
}

// RespAsyncRet 接收响应返回值的异步调用结果
type RespAsyncRet chan runtime.Ret

// Push 填入返回结果
func (resp RespAsyncRet) Push(ret Ret[any]) (retErr error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			retErr = panicErr
		}
	}()

	resp <- runtime.MakeRet(ret.Value, ret.Error)
	close(resp)

	return nil
}

// Cast 转换为异步调用结果
func (resp RespAsyncRet) Cast() runtime.AsyncRet {
	return chan runtime.Ret(resp)
}
