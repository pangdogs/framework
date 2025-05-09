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
	"reflect"
)

type ICallee interface {
	Callee(method string) reflect.Value
}

var (
	callChainRT = reflect.TypeFor[rpcstack.CallChain]()
)

func CallService(svcCtx service.Context, cc rpcstack.CallChain, addInName, method string, args variant.Array) (_ variant.Array, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("rpc: %w: %w", core.ErrPanicked, panicErr)
		}
	}()

	var scriptRV reflect.Value

	if addInName == "" {
		scriptRV = service.UnsafeContext(svcCtx).GetReflected()
	} else {
		ps, ok := svcCtx.GetAddInManager().Get(addInName)
		if !ok {
			return nil, ErrAddInNotFound
		}

		if ps.State() != extension.AddInState_Active {
			return nil, ErrAddInInactive
		}

		scriptRV = ps.Reflected()
	}

	methodRV := scriptRV.MethodByName(method)
	if !methodRV.IsValid() {
		callee, ok := scriptRV.Interface().(ICallee)
		if !ok {
			return nil, ErrMethodNotFound
		}

		methodRV = callee.Callee(method)
		if !methodRV.IsValid() {
			return nil, ErrMethodNotFound
		}
	}

	argsRV, err := parseArgs(methodRV, cc, args)
	if err != nil {
		return nil, err
	}

	return variant.MakeSerializedArray(methodRV.Call(argsRV))
}

func CallRuntime(svcCtx service.Context, cc rpcstack.CallChain, entityId uid.Id, addInName, method string, args variant.Array) (_ async.AsyncRet, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("rpc: %w: %w", core.ErrPanicked, panicErr)
		}
	}()

	return svcCtx.CallAsync(entityId, func(entity ec.Entity, _ ...any) async.Ret {
		var scriptRV reflect.Value

		if addInName == "" {
			scriptRV = runtime.UnsafeContext(runtime.Current(entity)).GetReflected()
		} else {
			ps, ok := runtime.Current(entity).GetAddInManager().Get(addInName)
			if !ok {
				return async.MakeRet(nil, ErrAddInNotFound)
			}

			if ps.State() != extension.AddInState_Active {
				return async.MakeRet(nil, ErrAddInInactive)
			}

			scriptRV = ps.Reflected()
		}

		methodRV := scriptRV.MethodByName(method)
		if !methodRV.IsValid() {
			callee, ok := scriptRV.Interface().(ICallee)
			if !ok {
				return async.MakeRet(nil, ErrMethodNotFound)
			}

			methodRV = callee.Callee(method)
			if !methodRV.IsValid() {
				return async.MakeRet(nil, ErrMethodNotFound)
			}
		}

		argsRV, err := parseArgs(methodRV, cc, args)
		if err != nil {
			return async.MakeRet(nil, err)
		}

		stack := rpcstack.Using(runtime.Current(entity))
		rpcstack.UnsafeRPCStack(stack).PushCallChain(cc)
		defer rpcstack.UnsafeRPCStack(stack).PopCallChain()

		retsRV := methodRV.Call(argsRV)
		if len(retsRV) == 1 {
			if asyncRet, ok := retsRV[0].Interface().(async.AsyncRet); ok {
				return async.MakeRet(asyncRet, nil)
			}
		}

		return async.MakeRet(variant.MakeSerializedArray(retsRV))
	}), nil
}

func CallEntity(svcCtx service.Context, cc rpcstack.CallChain, entityId uid.Id, component, method string, args variant.Array) (_ async.AsyncRet, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("rpc: %w: %w", core.ErrPanicked, panicErr)
		}
	}()

	return svcCtx.CallAsync(entityId, func(entity ec.Entity, _ ...any) async.Ret {
		var scriptRV reflect.Value

		if component == "" {
			scriptRV = entity.GetReflected()
		} else {
			comp := entity.GetComponent(component)
			if comp == nil {
				return async.MakeRet(nil, ErrComponentNotFound)
			}
			scriptRV = comp.GetReflected()
		}

		methodRV := scriptRV.MethodByName(method)
		if !methodRV.IsValid() {
			callee, ok := scriptRV.Interface().(ICallee)
			if !ok {
				return async.MakeRet(nil, ErrMethodNotFound)
			}

			methodRV = callee.Callee(method)
			if !methodRV.IsValid() {
				return async.MakeRet(nil, ErrMethodNotFound)
			}
		}

		argsRV, err := parseArgs(methodRV, cc, args)
		if err != nil {
			return async.MakeRet(nil, err)
		}

		stack := rpcstack.Using(runtime.Current(entity))
		rpcstack.UnsafeRPCStack(stack).PushCallChain(cc)
		defer rpcstack.UnsafeRPCStack(stack).PopCallChain()

		retsRV := methodRV.Call(argsRV)
		if len(retsRV) == 1 {
			if asyncRet, ok := retsRV[0].Interface().(async.AsyncRet); ok {
				return async.MakeRet(asyncRet, nil)
			}
		}

		return async.MakeRet(variant.MakeSerializedArray(retsRV))
	}), nil
}

func parseArgs(methodRV reflect.Value, cc rpcstack.CallChain, args variant.Array) ([]reflect.Value, error) {
	methodRT := methodRV.Type()
	var argsRV []reflect.Value
	var argsPos int

	switch methodRT.NumIn() {
	case len(args) + 1:
		if !callChainRT.AssignableTo(methodRT.In(0)) {
			return nil, ErrMethodParameterTypeMismatch
		}
		argsRV = append(make([]reflect.Value, 0, len(args)+1), reflect.ValueOf(cc))
		argsPos = 1

	case len(args):
		argsRV = make([]reflect.Value, 0, len(args))
		argsPos = 0

	default:
		return nil, ErrMethodParameterCountMismatch
	}

	for i := range args {
		argRV, err := args[i].Convert(methodRT.In(argsPos + i))
		if err != nil {
			return nil, ErrMethodParameterTypeMismatch
		}
		argsRV = append(argsRV, argRV)
	}

	return argsRV, nil
}

func waitAsyncRet(ctx context.Context, asyncRet async.AsyncRet) (variant.Array, error) {
	for {
		ret := asyncRet.Wait(ctx)
		if !ret.OK() {
			return nil, ret.Error
		}

		var ok bool
		asyncRet, ok = ret.Value.(async.AsyncRet)
		if ok {
			if asyncRet == nil {
				return nil, ErrAsyncMethodReturnedNil
			}
			continue
		}

		if rets, ok := ret.Value.(variant.Array); ok {
			return rets, nil
		}

		rets, err := variant.MakeSerializedArray([]any{ret.Value})
		if err != nil {
			return nil, err
		}

		return rets, nil
	}
}
