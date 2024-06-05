package rpcpcsr

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/rpcstack"
	"reflect"
	"slices"
)

func CallService(servCtx service.Context, callChain rpcstack.CallChain, plugin, method string, args variant.Array) (rets []reflect.Value, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()

	var reflected reflect.Value

	if plugin == "" {
		reflected = service.UnsafeContext(servCtx).GetReflected()
	} else {
		pi, ok := servCtx.GetPluginBundle().Get(plugin)
		if !ok {
			return nil, ErrPluginNotFound
		}
		reflected = pi.Reflected
	}

	methodRV := reflected.MethodByName(method)
	if !methodRV.IsValid() {
		return nil, ErrMethodNotFound
	}

	argsRV, err := prepareArgsRV(methodRV, callChain, args)
	if err != nil {
		return nil, err
	}

	return methodRV.Call(argsRV), nil
}

func CallRuntime(servCtx service.Context, callChain rpcstack.CallChain, entityId uid.Id, plugin, method string, args variant.Array) (asyncRet async.AsyncRet, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()

	return servCtx.Call(entityId, func(entity ec.Entity, a ...any) async.Ret {
		callChain := a[0].(rpcstack.CallChain)
		plugin := a[1].(string)
		method := a[2].(string)
		args := a[3].(variant.Array)

		var reflected reflect.Value

		if plugin == "" {
			reflected = runtime.UnsafeContext(runtime.Current(entity)).GetReflected()
		} else {
			pi, ok := runtime.Current(entity).GetPluginBundle().Get(plugin)
			if !ok {
				return async.MakeRet(nil, ErrPluginNotFound)
			}
			reflected = pi.Reflected
		}

		methodRV := reflected.MethodByName(method)
		if !methodRV.IsValid() {
			return async.MakeRet(nil, ErrMethodNotFound)
		}

		argsRV, err := prepareArgsRV(methodRV, callChain, args)
		if err != nil {
			return async.MakeRet(nil, err)
		}

		stack := rpcstack.Using(runtime.Current(entity))
		rpcstack.UnsafeRPCStack(stack).PushCallChain(callChain)
		defer rpcstack.UnsafeRPCStack(stack).PopCallChain()

		return async.MakeRet(methodRV.Call(argsRV), nil)
	}, callChain, plugin, method, args), nil
}

func CallEntity(servCtx service.Context, callChain rpcstack.CallChain, entityId uid.Id, component, method string, args variant.Array) (asyncRet async.AsyncRet, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()

	return servCtx.Call(entityId, func(entity ec.Entity, a ...any) async.Ret {
		callChain := a[0].(rpcstack.CallChain)
		compName := a[1].(string)
		method := a[2].(string)
		args := a[3].(variant.Array)

		var reflected reflect.Value

		if compName == "" {
			reflected = ec.UnsafeEntity(entity).GetReflected()
		} else {
			comp := entity.GetComponent(compName)
			if comp == nil {
				return async.MakeRet(nil, ErrComponentNotFound)
			}
			reflected = ec.UnsafeComponent(comp).GetReflected()
		}

		methodRV := reflected.MethodByName(method)
		if !methodRV.IsValid() {
			return async.MakeRet(nil, ErrMethodNotFound)
		}

		argsRV, err := prepareArgsRV(methodRV, callChain, args)
		if err != nil {
			return async.MakeRet(nil, err)
		}

		stack := rpcstack.Using(runtime.Current(entity))
		rpcstack.UnsafeRPCStack(stack).PushCallChain(callChain)
		defer rpcstack.UnsafeRPCStack(stack).PopCallChain()

		return async.MakeRet(methodRV.Call(argsRV), nil)
	}, callChain, component, method, args), nil
}

var (
	callChainRT = reflect.TypeFor[rpcstack.CallChain]()
)

func prepareArgsRV(methodRV reflect.Value, callChain rpcstack.CallChain, args variant.Array) ([]reflect.Value, error) {
	methodRT := methodRV.Type()
	var argsRV []reflect.Value
	var argsPos int

	switch methodRT.NumIn() {
	case len(args) + 1:
		if !callChainRT.AssignableTo(methodRT.In(0)) {
			return nil, ErrMethodParameterTypeMismatch
		}
		argsRV = append(make([]reflect.Value, 0, len(args)+1), reflect.ValueOf(slices.Clone(callChain)))
		argsPos = 1

	case len(args):
		argsRV = make([]reflect.Value, 0, len(args))
		argsPos = 0

	default:
		return nil, ErrMethodParameterCountMismatch
	}

	for i := range args {
		argRV := args[i].Reflected
		argRT := argRV.Type()
		paramRT := methodRT.In(argsPos + i)

	retry:
		if argRT.AssignableTo(paramRT) {
			argsRV = append(argsRV, argRV)
			continue
		}

		if argRV.CanConvert(paramRT) {
			if argRT.Size() > paramRT.Size() {
				return nil, ErrMethodParameterTypeMismatch
			}
			argsRV = append(argsRV, argRV.Convert(paramRT))
			continue
		}

		if argRT.Kind() == reflect.Pointer {
			argRV = argRV.Elem()
			argRT = argRV.Type()
			goto retry
		}

		argRV, err := variant.CastVariantReflected(args[i], paramRT)
		if err != nil {
			return nil, ErrMethodParameterTypeMismatch
		}

		argsRV = append(argsRV, argRV)
	}

	return argsRV, nil
}
