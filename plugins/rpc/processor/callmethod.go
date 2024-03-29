package processor

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/net/gap/variant"
	"reflect"
)

func CallService(servCtx service.Context, plugin, method string, args variant.Array) (rets []reflect.Value, err error) {
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

	argsRV, err := prepareArgsRV(methodRV, args)
	if err != nil {
		return nil, err
	}

	return methodRV.Call(argsRV), nil
}

func CallRuntime(servCtx service.Context, entityId uid.Id, plugin, method string, args variant.Array) (asyncRet runtime.AsyncRet, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()

	return servCtx.Call(entityId, func(entity ec.Entity, a ...any) service.Ret {
		plugin := a[0].(string)
		method := a[1].(string)
		args := a[2].(variant.Array)

		var reflected reflect.Value

		if plugin == "" {
			reflected = runtime.UnsafeContext(runtime.Current(entity)).GetReflected()
		} else {
			pi, ok := runtime.Current(entity).GetPluginBundle().Get(plugin)
			if !ok {
				return runtime.MakeRet(nil, ErrPluginNotFound)
			}
			reflected = pi.Reflected
		}

		methodRV := reflected.MethodByName(method)
		if !methodRV.IsValid() {
			return runtime.MakeRet(nil, ErrMethodNotFound)
		}

		argsRV, err := prepareArgsRV(methodRV, args)
		if err != nil {
			return runtime.MakeRet(nil, err)
		}

		return runtime.MakeRet(methodRV.Call(argsRV), nil)
	}, plugin, method, args), nil
}

func CallEntity(servCtx service.Context, entityId uid.Id, component, method string, args variant.Array) (asyncRet runtime.AsyncRet, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()

	return servCtx.Call(entityId, func(entity ec.Entity, a ...any) service.Ret {
		compName := a[0].(string)
		method := a[1].(string)
		args := a[2].(variant.Array)

		comp := entity.GetComponent(compName)
		if comp == nil {
			return runtime.MakeRet(nil, ErrComponentNotFound)
		}

		methodRV := ec.UnsafeComponent(comp).GetReflected().MethodByName(method)
		if !methodRV.IsValid() {
			return runtime.MakeRet(nil, ErrMethodNotFound)
		}

		argsRV, err := prepareArgsRV(methodRV, args)
		if err != nil {
			return runtime.MakeRet(nil, err)
		}

		return runtime.MakeRet(methodRV.Call(argsRV), nil)
	}, component, method, args), nil
}

func prepareArgsRV(methodRV reflect.Value, args variant.Array) ([]reflect.Value, error) {
	methodRT := methodRV.Type()
	if methodRT.NumIn() != len(args) {
		return nil, ErrMethodParameterCountMismatch
	}

	argsRV := make([]reflect.Value, 0, len(args))

	for i := range args {
		argRV := args[i].Reflected
		argRT := argRV.Type()
		paramRT := methodRT.In(i)

	retry:
		if !argRT.AssignableTo(paramRT) {
			if argRV.CanConvert(paramRT) {
				if argRT.Size() > paramRT.Size() {
					return nil, ErrMethodParameterTypeMismatch
				}
				argRV = argRV.Convert(paramRT)
			} else {
				if argRT.Kind() != reflect.Pointer {
					return nil, ErrMethodParameterTypeMismatch
				}
				argRV = argRV.Elem()
				argRT = argRV.Type()
				goto retry
			}
		}

		argsRV = append(argsRV, argRV)
	}

	return argsRV, nil
}
