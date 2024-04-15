package concurrent

import (
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/types"
)

// RespFunc 接收响应返回值的函数
type RespFunc[T any] generic.Action1[Ret[T]]

// Push 填入返回结果
func (resp RespFunc[T]) Push(ret Ret[any]) error {
	if !ret.OK() {
		generic.MakeAction1(resp).Exec(MakeRet[T](types.Zero[T](), ret.Error))
		return nil
	}

	v, ok := ret.Value.(T)
	if !ok {
		generic.MakeAction1(resp).Exec(MakeRet[T](types.Zero[T](), ErrFutureRespIncorrectType))
		return nil
	}

	generic.MakeAction1(resp).Exec(MakeRet[T](v, nil))
	return nil
}

// RespDelegate 接收响应返回值的委托
type RespDelegate[T any] generic.DelegateAction1[Ret[T]]

// Push 填入返回结果
func (resp RespDelegate[T]) Push(ret Ret[any]) error {
	if !ret.OK() {
		generic.DelegateAction1[Ret[T]](resp).Exec(nil, MakeRet[T](types.Zero[T](), ret.Error))
		return nil
	}

	v, ok := ret.Value.(T)
	if !ok {
		generic.DelegateAction1[Ret[T]](resp).Exec(nil, MakeRet[T](types.Zero[T](), ErrFutureRespIncorrectType))
		return nil
	}

	generic.DelegateAction1[Ret[T]](resp).Exec(nil, MakeRet[T](v, nil))
	return nil
}
