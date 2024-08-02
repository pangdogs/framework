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

package framework

import (
	"context"
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
)

// AwaitDirector 异步等待分发器
type AwaitDirector struct {
	iec      iEC
	director core.AwaitDirector
}

// Any 异步等待任意一个结果返回
func (ad AwaitDirector) Any(fun generic.ActionVar1[async.Ret, any], va ...any) {
	ad.director.Any(func(_ runtime.Context, ret async.Ret, a ...any) {
		if !ad.iec.IsAlive() {
			return
		}
		fun.Exec(ret, a...)
	}, va...)
}

// AnyOK 异步等待任意一个结果成功返回
func (ad AwaitDirector) AnyOK(fun generic.ActionVar1[async.Ret, any], va ...any) {
	ad.director.AnyOK(func(_ runtime.Context, ret async.Ret, a ...any) {
		if !ad.iec.IsAlive() {
			return
		}
		fun.Exec(ret, a...)
	}, va...)
}

// All 异步等待所有结果返回
func (ad AwaitDirector) All(fun generic.ActionVar1[[]async.Ret, any], va ...any) {
	ad.director.All(func(_ runtime.Context, rets []async.Ret, a ...any) {
		if !ad.iec.IsAlive() {
			return
		}
		fun.Exec(rets, a...)
	}, va...)
}

// Pipe 异步等待管道返回
func (ad AwaitDirector) Pipe(ctx context.Context, fun generic.ActionVar1[async.Ret, any], va ...any) {
	ad.director.Pipe(ctx, func(_ runtime.Context, ret async.Ret, a ...any) {
		if !ad.iec.IsAlive() {
			return
		}
		fun.Exec(ret, a...)
	}, va...)
}
