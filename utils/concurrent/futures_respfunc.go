package concurrent

import (
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/types"
)

// RespFunc 接收响应返回值的函数
type RespFunc[T any] generic.Action1[async.RetT[T]]

// Push 填入返回结果
func (fun RespFunc[T]) Push(ret async.RetT[any]) error {
	if !ret.OK() {
		generic.MakeAction1(fun).Exec(async.MakeRetT[T](types.ZeroT[T](), ret.Error))
		return nil
	}

	resp, ok := async.AsRetT[T](ret)
	if !ok {
		generic.MakeAction1(fun).Exec(async.MakeRetT[T](types.ZeroT[T](), ErrFutureRespIncorrectType))
		return nil
	}

	generic.MakeAction1(fun).Exec(resp)
	return nil
}

// RespDelegate 接收响应返回值的委托
type RespDelegate[T any] generic.DelegateAction1[async.RetT[T]]

// Push 填入返回结果
func (dlg RespDelegate[T]) Push(ret async.RetT[any]) error {
	if !ret.OK() {
		generic.DelegateAction1[async.RetT[T]](dlg).Exec(nil, async.MakeRetT[T](types.ZeroT[T](), ret.Error))
		return nil
	}

	resp, ok := async.AsRetT[T](ret)
	if !ok {
		generic.DelegateAction1[async.RetT[T]](dlg).Exec(nil, async.MakeRetT[T](types.ZeroT[T](), ErrFutureRespIncorrectType))
		return nil
	}

	generic.DelegateAction1[async.RetT[T]](dlg).Exec(nil, resp)
	return nil
}
