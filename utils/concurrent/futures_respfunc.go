/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

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
		generic.CastAction1(fun).Exec(async.MakeRetT[T](types.ZeroT[T](), ret.Error))
		return nil
	}

	resp, ok := async.AsRetT[T](ret)
	if !ok {
		generic.CastAction1(fun).Exec(async.MakeRetT[T](types.ZeroT[T](), ErrFutureRespIncorrectType))
		return nil
	}

	generic.CastAction1(fun).Exec(resp)
	return nil
}

// RespDelegate 接收响应返回值的委托
type RespDelegate[T any] generic.DelegateVoid1[async.RetT[T]]

// Push 填入返回结果
func (dlg RespDelegate[T]) Push(ret async.RetT[any]) error {
	if !ret.OK() {
		generic.DelegateVoid1[async.RetT[T]](dlg).Exec(nil, async.MakeRetT[T](types.ZeroT[T](), ret.Error))
		return nil
	}

	resp, ok := async.AsRetT[T](ret)
	if !ok {
		generic.DelegateVoid1[async.RetT[T]](dlg).Exec(nil, async.MakeRetT[T](types.ZeroT[T](), ErrFutureRespIncorrectType))
		return nil
	}

	generic.DelegateVoid1[async.RetT[T]](dlg).Exec(nil, resp)
	return nil
}
