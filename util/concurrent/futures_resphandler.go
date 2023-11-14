package concurrent

import (
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/golaxy/util/types"
)

// RespHandler 接收响应返回值的处理器
type RespHandler[T any] generic.Action1[Ret[T]]

// Push 填入返回结果
func (resp RespHandler[T]) Push(ret Ret[any]) error {
	if !ret.OK() {
		generic.CastAction1(resp).Exec(MakeRet[T](types.Zero[T](), ret.Error))
		return nil
	}

	v, ok := ret.Value.(T)
	if !ok {
		generic.CastAction1(resp).Exec(MakeRet[T](types.Zero[T](), ErrFutureRespIncorrectType))
		return nil
	}

	generic.CastAction1(resp).Exec(MakeRet[T](v, ret.Error))
	return nil
}
