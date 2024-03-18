package processor

import (
	"context"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/types"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/variant"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/rpc/callpath"
	"git.golaxy.org/framework/util/concurrent"
	"reflect"
)

var (
	ErrPluginNotFound               = errors.New("rpc: plugin not found")
	ErrMethodNotFound               = errors.New("rpc: method not found")
	ErrComponentNotFound            = errors.New("rpc: component not found")
	ErrMethodParameterCountMismatch = errors.New("rpc: method parameter count mismatch")
	ErrMethodParameterTypeMismatch  = errors.New("rpc: method parameter type mismatch")
)

// NewServiceDispatcher 创建分布式服务间的RPC分发器
func NewServiceDispatcher() IDispatcher {
	return &_ServiceDispatcher{}
}

// _ServiceDispatcher 分布式服务间的RPC分发器
type _ServiceDispatcher struct {
	servCtx service.Context
	dist    dserv.IDistService
	watcher dserv.IWatcher
}

// Init 初始化
func (d *_ServiceDispatcher) Init(ctx service.Context) {
	d.servCtx = ctx
	d.dist = dserv.Using(ctx)
	d.watcher = d.dist.WatchMsg(context.Background(), generic.CastDelegateFunc2(d.handleMsg))

	log.Debugf(d.servCtx, "rpc dispatcher %q started", types.AnyFullName(*d))
}

// Shut 结束
func (d *_ServiceDispatcher) Shut(ctx service.Context) {
	<-d.watcher.Stop()

	log.Debugf(d.servCtx, "rpc dispatcher %q stopped", types.AnyFullName(*d))
}

func (d *_ServiceDispatcher) handleMsg(topic string, mp gap.MsgPacket) error {
	addr := d.dist.GetAddressDetails()

	if !addr.InDomain(mp.Head.Src) {
		return nil
	}

	switch mp.Head.MsgId {
	case gap.MsgId_OneWayRPC:
		return d.acceptNotify(mp.Msg.(*gap.MsgOneWayRPC))

	case gap.MsgId_RPC_Request:
		return d.acceptRequest(mp.Head.Src, mp.Msg.(*gap.MsgRPCRequest))

	case gap.MsgId_RPC_Reply:
		return d.resolve(mp.Msg.(*gap.MsgRPCReply))
	}

	return nil
}

func (d *_ServiceDispatcher) acceptNotify(req *gap.MsgOneWayRPC) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		return fmt.Errorf("parse rpc notify path:%q failed, %s", req.Path, err)
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			if _, err := d.callService(cp.Plugin, cp.Method, req.Args); err != nil {
				log.Errorf(d.servCtx, "rpc notify service plugin:%q, method:%q calls failed, %s", cp.Plugin, cp.Method, err)
				return
			}
			log.Debugf(d.servCtx, "rpc notify service plugin:%q, method:%q calls finished", cp.Plugin, cp.Method)
			return
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := d.callRuntime(uid.From(cp.EntityId), cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls failed, %s", cp.EntityId, cp.Plugin, cp.Method, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls failed, %s", cp.EntityId, cp.Plugin, cp.Method, ret.Error)
				return
			}
			log.Debugf(d.servCtx, "rpc notify entity:%q, runtime plugin:%q, method:%q calls finished", cp.EntityId, cp.Plugin, cp.Method)
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := d.callEntity(uid.From(cp.EntityId), cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc notify entity:%q, component:%q, method:%q calls failed, %s", cp.EntityId, cp.Component, cp.Method, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc notify entity:%q, component:%q, method:%q calls failed, %s", cp.EntityId, cp.Component, cp.Method, ret.Error)
				return
			}
			log.Debugf(d.servCtx, "rpc notify entity:%q, component:%q, method:%q calls finished", cp.EntityId, cp.Component, cp.Method)
		}()

		return nil
	}

	return nil
}

func (d *_ServiceDispatcher) acceptRequest(src string, req *gap.MsgRPCRequest) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		err = fmt.Errorf("parse rpc request(%d) path %q failed, %s", req.CorrId, req.Path, err)
		go d.reply(src, req.CorrId, nil, err)
		return err
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			retsRV, err := d.callService(cp.Plugin, cp.Method, req.Args)
			if err != nil {
				log.Errorf(d.servCtx, "rpc request(%d) service plugin:%q, method:%q calls failed, %s", req.CorrId, cp.Plugin, cp.Method, err)
				d.reply(src, req.CorrId, nil, err)
				return
			}
			log.Debugf(d.servCtx, "rpc request(%d) service plugin:%q, method:%q calls finished", req.CorrId, cp.Plugin, cp.Method)
			d.reply(src, req.CorrId, retsRV, nil)
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := d.callRuntime(uid.From(cp.EntityId), cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, err)
			go d.reply(src, req.CorrId, nil, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, ret.Error)
				d.reply(src, req.CorrId, nil, ret.Error)
				return
			}
			log.Debugf(d.servCtx, "rpc request(%d) entity:%q, runtime plugin:%q, method:%q calls finished", req.CorrId, cp.EntityId, cp.Plugin, cp.Method)
			d.reply(src, req.CorrId, ret.Value.([]reflect.Value), nil)
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := d.callEntity(uid.From(cp.EntityId), cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Component, cp.Method, err)
			go d.reply(src, req.CorrId, nil, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls failed, %s", req.CorrId, cp.EntityId, cp.Component, cp.Method, ret.Error)
				d.reply(src, req.CorrId, nil, ret.Error)
				return
			}
			log.Debugf(d.servCtx, "rpc request(%d) entity:%q, component:%q, method:%q calls finished", req.CorrId, cp.EntityId, cp.Component, cp.Method)
			d.reply(src, req.CorrId, ret.Value.([]reflect.Value), nil)
		}()

		return nil
	}

	return nil
}

func (d *_ServiceDispatcher) callService(plugin, method string, args variant.Array) (rets []reflect.Value, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()

	var reflected reflect.Value

	if plugin == "" {
		reflected = service.UnsafeContext(d.servCtx).GetReflected()
	} else {
		pi, ok := d.servCtx.GetPluginBundle().Get(plugin)
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

func (d *_ServiceDispatcher) callRuntime(entityId uid.Id, plugin, method string, args variant.Array) (asyncRet runtime.AsyncRet, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()

	return d.servCtx.Call(entityId, func(entity ec.Entity, a ...any) service.Ret {
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

func (d *_ServiceDispatcher) callEntity(entityId uid.Id, component, method string, args variant.Array) (asyncRet runtime.AsyncRet, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()

	return d.servCtx.Call(entityId, func(entity ec.Entity, a ...any) service.Ret {
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

func (d *_ServiceDispatcher) reply(src string, corrId int64, retsRV []reflect.Value, retErr error) {
	msg := &gap.MsgRPCReply{
		CorrId: corrId,
	}

	var err error
	msg.Rets, err = variant.MakeArray(retsRV)
	if err != nil {
		log.Errorf(d.servCtx, "rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}

	if retErr != nil {
		msg.Error = *variant.MakeError(retErr)
	}

	err = d.dist.SendMsg(src, msg)
	if err != nil {
		log.Errorf(d.servCtx, "rpc reply(%d) to src:%q failed, %s", corrId, src, err)
		return
	}

	log.Debugf(d.servCtx, "rpc reply(%d) to src:%q ok", corrId, src)
}

func (d *_ServiceDispatcher) resolve(reply *gap.MsgRPCReply) error {
	ret := concurrent.Ret[any]{}

	if reply.Error.OK() {
		if len(reply.Rets) > 0 {
			ret.Value = reply.Rets
		}
	} else {
		ret.Error = &reply.Error
	}

	return d.dist.GetFutures().Resolve(reply.CorrId, ret)
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
