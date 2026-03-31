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

type ICallee interface {
	Callee(method string) reflect.Value
}

var (
	callChainRT = reflect.TypeFor[rpcstack.CallChain]()
)

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
			return nil, ErrAddInNotFound
		}

		if status.State() != extension.AddInState_Running {
			return nil, ErrAddInInactive
		}

		scriptRV = status.Reflected()
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

	return variant.NewSerializedArray(methodRV.Call(argsRV))
}

func CallRuntime(svcCtx service.Context, cc rpcstack.CallChain, entityId uid.Id, addIn, method string, args variant.Array) (_ async.Future, err error) {
	return svcCtx.CallAsync(entityId, func(entity ec.Entity, _ ...any) async.Result {
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

		return async.NewResult(variant.NewSerializedArray(retsRV))
	}), nil
}

func CallEntity(svcCtx service.Context, cc rpcstack.CallChain, entityId uid.Id, component, method string, args variant.Array) (_ async.Future, err error) {
	return svcCtx.CallAsync(entityId, func(entity ec.Entity, _ ...any) async.Result {
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

		return async.NewResult(variant.NewSerializedArray(retsRV))
	}), nil
}

func parseArgs(methodRV reflect.Value, cc rpcstack.CallChain, args variant.Array) ([]reflect.Value, error) {
	methodRT := methodRV.Type()
	ccPos := -1

	for i := range methodRT.NumIn() {
		if !callChainRT.AssignableTo(methodRT.In(i)) {
			continue
		}
		if ccPos >= 0 {
			return nil, ErrMethodParameterCountMismatch
		}
		ccPos = i
	}

	switch {
	case ccPos < 0 && methodRT.NumIn() != len(args):
		return nil, ErrMethodParameterCountMismatch
	case ccPos >= 0 && methodRT.NumIn() != len(args)+1:
		return nil, ErrMethodParameterCountMismatch
	}

	argsRV := make([]reflect.Value, methodRT.NumIn())
	j := 0

	for i := range argsRV {
		if i == ccPos {
			argsRV[i] = reflect.ValueOf(cc)
			continue
		}
		if j >= len(args) {
			return nil, ErrMethodParameterCountMismatch
		}

		argRV, err := args[j].Convert(methodRT.In(i))
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
			return nil, ret.Error
		}

		var ok bool
		future, ok = ret.Value.(async.Future)
		if ok {
			if future.IsNil() {
				return nil, ErrAsyncMethodReturnedNil
			}
			continue
		}

		if rets, ok := ret.Value.(variant.Array); ok {
			return rets, nil
		}

		rets, err := variant.NewSerializedArray([]any{ret.Value})
		if err != nil {
			return nil, err
		}

		return rets, nil
	}
}
