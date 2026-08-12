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

package rpcpcsr

import (
	"context"
	"fmt"
	"reflect"

	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/extension"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/rpcstack"
	"git.golaxy.org/framework/net/gap/variant"
)

// ICallee 允许对象按名称动态提供可调用方法。
type ICallee interface {
	// Callee 返回指定名称的方法；无对应方法时返回无效 reflect.Value。
	Callee(method string) reflect.Value
}

var (
	callChainRT = reflect.TypeFor[rpcstack.CallChain]()
)

// CallService 在当前服务或其运行中的插件上同步调用方法，并将 panic 转换为错误。
func CallService(svcCtx service.Context, cc rpcstack.CallChain, addIn, method string, args variant.Array) (_ variant.Array, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("rpc: %w: %w", core.ErrPanicked, panicErr)
		}
	}()

	var scriptRV reflect.Value

	if addIn == "" {
		scriptRV = service.UnsafeContext(svcCtx).Reflected()
	} else {
		status, ok := svcCtx.AddInManager().GetStatusByName(addIn)
		if !ok {
			return variant.Array{}, ErrAddInNotFound
		}

		if status.State() != extension.AddInState_Running {
			return variant.Array{}, ErrAddInInactive
		}

		scriptRV = status.Reflected()
	}

	methodRV := scriptRV.MethodByName(method)
	if !methodRV.IsValid() {
		callee, ok := scriptRV.Interface().(ICallee)
		if !ok {
			return variant.Array{}, ErrMethodNotFound
		}

		methodRV = callee.Callee(method)
		if !methodRV.IsValid() {
			return variant.Array{}, ErrMethodNotFound
		}
	}

	argsRV, err := parseArgs(methodRV, cc, args)
	if err != nil {
		return variant.Array{}, err
	}

	return variant.NewArray(methodRV.Call(argsRV))
}

// CallRuntime 将方法调用调度到实体所在的运行时；addIn 为空时调用运行时本身。
func CallRuntime(svcCtx service.Context, cc rpcstack.CallChain, entityId uid.Id, addIn, method string, args variant.Array) (_ async.Future, err error) {
	return svcCtx.Submit(entityId, func(entity ec.Entity, _ ...any) async.Result {
		var scriptRV reflect.Value

		if addIn == "" {
			scriptRV = runtime.UnsafeContext(runtime.Current(entity)).Reflected()
		} else {
			status, ok := runtime.Current(entity).AddInManager().GetStatusByName(addIn)
			if !ok {
				return async.NewResult(nil, ErrAddInNotFound)
			}

			if status.State() != extension.AddInState_Running {
				return async.NewResult(nil, ErrAddInInactive)
			}

			scriptRV = status.Reflected()
		}

		methodRV := scriptRV.MethodByName(method)
		if !methodRV.IsValid() {
			callee, ok := scriptRV.Interface().(ICallee)
			if !ok {
				return async.NewResult(nil, ErrMethodNotFound)
			}

			methodRV = callee.Callee(method)
			if !methodRV.IsValid() {
				return async.NewResult(nil, ErrMethodNotFound)
			}
		}

		argsRV, err := parseArgs(methodRV, cc, args)
		if err != nil {
			return async.NewResult(nil, err)
		}

		stack := rpcstack.AddIn.Require(runtime.Current(entity))
		rpcstack.UnsafeRPCStack(stack).PushCallChain(cc)
		defer rpcstack.UnsafeRPCStack(stack).PopCallChain()

		retsRV := methodRV.Call(argsRV)
		if len(retsRV) == 1 {
			if future, ok := retsRV[0].Interface().(async.Future); ok {
				return async.NewResult(future, nil)
			}
		}

		rets, err := variant.NewArray(retsRV)
		if err != nil {
			return async.NewResult(nil, err)
		}

		return async.NewResult(rets.Snapshot(true))
	}), nil
}

// CallEntity 将方法调用调度到实体；component 为空时调用实体本身。
func CallEntity(svcCtx service.Context, cc rpcstack.CallChain, entityId uid.Id, component, method string, args variant.Array) (_ async.Future, err error) {
	return svcCtx.Submit(entityId, func(entity ec.Entity, _ ...any) async.Result {
		var scriptRV reflect.Value

		if component == "" {
			scriptRV = entity.Reflected()
		} else {
			comp := entity.GetComponent(component)
			if comp == nil {
				return async.NewResult(nil, ErrComponentNotFound)
			}
			scriptRV = comp.Reflected()
		}

		methodRV := scriptRV.MethodByName(method)
		if !methodRV.IsValid() {
			callee, ok := scriptRV.Interface().(ICallee)
			if !ok {
				return async.NewResult(nil, ErrMethodNotFound)
			}

			methodRV = callee.Callee(method)
			if !methodRV.IsValid() {
				return async.NewResult(nil, ErrMethodNotFound)
			}
		}

		argsRV, err := parseArgs(methodRV, cc, args)
		if err != nil {
			return async.NewResult(nil, err)
		}

		stack := rpcstack.AddIn.Require(runtime.Current(entity))
		rpcstack.UnsafeRPCStack(stack).PushCallChain(cc)
		defer rpcstack.UnsafeRPCStack(stack).PopCallChain()

		retsRV := methodRV.Call(argsRV)
		if len(retsRV) == 1 {
			if future, ok := retsRV[0].Interface().(async.Future); ok {
				return async.NewResult(future, nil)
			}
		}

		rets, err := variant.NewArray(retsRV)
		if err != nil {
			return async.NewResult(nil, err)
		}

		return async.NewResult(rets.Snapshot(true))
	}), nil
}

func parseArgs(methodRV reflect.Value, cc rpcstack.CallChain, args variant.Array) ([]reflect.Value, error) {
	methodRT := methodRV.Type()
	ccPos := -1

	for i := range methodRT.NumIn() {
		if methodRT.In(i) != callChainRT {
			continue
		}
		if ccPos >= 0 {
			return nil, ErrMethodParameterCountMismatch
		}
		ccPos = i
	}

	switch {
	case ccPos < 0 && methodRT.NumIn() != len(args.Items):
		return nil, ErrMethodParameterCountMismatch
	case ccPos >= 0 && methodRT.NumIn() != len(args.Items)+1:
		return nil, ErrMethodParameterCountMismatch
	}

	argsRV := make([]reflect.Value, methodRT.NumIn())
	j := 0

	for i := range argsRV {
		if i == ccPos {
			argsRV[i] = reflect.ValueOf(cc)
			continue
		}
		if j >= len(args.Items) {
			return nil, ErrMethodParameterCountMismatch
		}

		argRV, err := args.Items[j].ToNative(methodRT.In(i))
		if err != nil {
			return nil, ErrMethodParameterTypeMismatch
		}

		argsRV[i] = argRV
		j++
	}

	return argsRV, nil
}

func waitAsyncResult(ctx context.Context, future async.Future) (variant.Array, error) {
	for {
		ret := future.Wait(ctx)
		if !ret.OK() {
			return variant.Array{}, ret.Error
		}

		var ok bool
		future, ok = ret.Value.(async.Future)
		if ok {
			if future.IsNil() {
				return variant.Array{}, ErrAsyncMethodReturnedNil
			}
			continue
		}

		if rets, ok := ret.Value.(variant.Array); ok {
			return rets, nil
		}

		rets, err := variant.NewArray([]any{ret.Value})
		if err != nil {
			return variant.Array{}, err
		}

		return rets.Snapshot(true)
	}
}
