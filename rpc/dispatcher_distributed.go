package rpc

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/ec"
	"kit.golaxy.org/golaxy/runtime"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/golaxy/util/types"
	"kit.golaxy.org/golaxy/util/uid"
	"kit.golaxy.org/plugins/distributed"
	"kit.golaxy.org/plugins/gap"
	"kit.golaxy.org/plugins/gap/variant"
	"kit.golaxy.org/plugins/log"
	"kit.golaxy.org/plugins/rpc/callpath"
	"kit.golaxy.org/plugins/util/concurrent"
	"reflect"
	"strings"
)

var (
	ErrPluginNotFound               = errors.New("rpc: plugin not found")
	ErrMethodNotFound               = errors.New("rpc: method not found")
	ErrComponentNotFound            = errors.New("rpc: component not found")
	ErrMethodParameterCountMismatch = errors.New("rpc: method parameter count mismatch")
	ErrMethodParameterTypeMismatch  = errors.New("rpc: method parameter type mismatch")
)

// DistributedDispatcher 分布式服务的RPC分发器
type DistributedDispatcher struct {
	servCtx service.Context
	dist    distributed.Distributed
	watcher distributed.Watcher
}

// Init 初始化
func (d *DistributedDispatcher) Init(ctx service.Context) {
	d.servCtx = ctx
	d.dist = distributed.Using(ctx)
	d.watcher = d.dist.WatchMsg(context.Background(), generic.CastDelegateFunc2(d.handleMsg))

	log.Debugf(d.servCtx, "dispatcher %q started", types.AnyFullName(*d))
}

// Shut 结束
func (d *DistributedDispatcher) Shut(ctx service.Context) {
	<-d.watcher.Stop()

	log.Debugf(d.servCtx, "dispatcher %q stopped", types.AnyFullName(*d))
}

func (d *DistributedDispatcher) handleMsg(topic string, mp gap.MsgPacket) error {
	addr := d.dist.GetAddress()

	if !strings.HasPrefix(mp.Head.Src, addr.Domain) {
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

func (d *DistributedDispatcher) acceptNotify(req *gap.MsgOneWayRPC) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		return fmt.Errorf("parse call path %q failed, %s", req.Path, err)
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			if _, err := d.callService(cp.Plugin, cp.Method, req.Args); err != nil {
				log.Errorf(d.servCtx, "service plugin %q method %q calls failed, %s", cp.Plugin, cp.Method, err)
				return
			}
			log.Debugf(d.servCtx, "service plugin %q method %q calls finished", cp.Plugin, cp.Method)
			return
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := d.callRuntime(uid.Id(cp.EntityId), cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "entity %q runtime plugin %q method %q calls failed, %s", cp.EntityId, cp.Plugin, cp.Method, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "entity %q runtime plugin %q method %q calls failed, %s", cp.EntityId, cp.Plugin, cp.Method, ret.Error)
				return
			}
			log.Debugf(d.servCtx, "entity %q runtime plugin %q method %q calls finished", cp.EntityId, cp.Plugin, cp.Method)
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := d.callEntity(uid.Id(cp.EntityId), cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "entity %q component %q method %q calls failed, %s", cp.EntityId, cp.Component, cp.Method, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "entity %q component %q method %q calls failed, %s", cp.EntityId, cp.Component, cp.Method, ret.Error)
				return
			}
			log.Debugf(d.servCtx, "entity %q component %q method %q calls finished", cp.EntityId, cp.Component, cp.Method)
		}()

		return nil
	}

	return nil
}

func (d *DistributedDispatcher) acceptRequest(src string, req *gap.MsgRPCRequest) error {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		err = fmt.Errorf("%d: parse call path %q failed, %s", req.CorrId, req.Path, err)
		go d.reply(src, req.CorrId, nil, err)
		return err
	}

	switch cp.Category {
	case callpath.Service:
		go func() {
			retsRV, err := d.callService(cp.Plugin, cp.Method, req.Args)
			if err != nil {
				log.Errorf(d.servCtx, "%d: service plugin %q method %q calls failed, %s", req.CorrId, cp.Plugin, cp.Method, err)
				d.reply(src, req.CorrId, nil, err)
				return
			}
			log.Debugf(d.servCtx, "%d: service plugin %q method %q calls finished", req.CorrId, cp.Plugin, cp.Method)
			d.reply(src, req.CorrId, retsRV, nil)
		}()

		return nil

	case callpath.Runtime:
		asyncRet, err := d.callRuntime(uid.Id(cp.EntityId), cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "%d: entity %q runtime plugin %q method %q calls failed, %s", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, err)
			go d.reply(src, req.CorrId, nil, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "%d: entity %q runtime plugin %q method %q calls failed, %s", req.CorrId, cp.EntityId, cp.Plugin, cp.Method, ret.Error)
				d.reply(src, req.CorrId, nil, ret.Error)
				return
			}
			log.Debugf(d.servCtx, "%d: entity %q runtime plugin %q method %q calls finished", req.CorrId, cp.EntityId, cp.Plugin, cp.Method)
			d.reply(src, req.CorrId, ret.Value.([]reflect.Value), nil)
		}()

		return nil

	case callpath.Entity:
		asyncRet, err := d.callEntity(uid.Id(cp.EntityId), cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.servCtx, "%d: entity %q component %q method %q calls failed, %s", req.CorrId, cp.EntityId, cp.Component, cp.Method, err)
			go d.reply(src, req.CorrId, nil, err)
			return nil
		}

		go func() {
			ret := asyncRet.Wait(d.servCtx)
			if !ret.OK() {
				log.Errorf(d.servCtx, "%d: entity %q component %q method %q calls failed, %s", req.CorrId, cp.EntityId, cp.Component, cp.Method, ret.Error)
				d.reply(src, req.CorrId, nil, ret.Error)
				return
			}
			log.Debugf(d.servCtx, "%d: entity %q component %q method %q calls finished", req.CorrId, cp.EntityId, cp.Component, cp.Method)
			d.reply(src, req.CorrId, ret.Value.([]reflect.Value), nil)
		}()

		return nil
	}

	return nil
}

func (d *DistributedDispatcher) callService(plugin, method string, args variant.Array) (rets []reflect.Value, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", golaxy.ErrPanicked, panicErr)
		}
	}()

	pi, ok := d.servCtx.GetPluginBundle().Get(plugin)
	if !ok {
		return nil, ErrPluginNotFound
	}

	methodRV := pi.Reflected.MethodByName(method)
	if methodRV.IsZero() {
		return nil, ErrMethodNotFound
	}

	argsRV, err := prepareArgsRV(methodRV, args)
	if err != nil {
		return nil, err
	}

	return methodRV.Call(argsRV), nil
}

func (d *DistributedDispatcher) callRuntime(entityId uid.Id, plugin, method string, args variant.Array) (asyncRet runtime.AsyncRet, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", golaxy.ErrPanicked, panicErr)
		}
	}()

	return d.servCtx.Call(entityId, func(entity ec.Entity, a ...any) service.Ret {
		plugin := a[0].(string)
		method := a[1].(string)
		args := a[2].(variant.Array)

		pi, ok := runtime.Current(entity).GetPluginBundle().Get(plugin)
		if !ok {
			return runtime.MakeRet(nil, ErrPluginNotFound)
		}

		methodRV := pi.Reflected.MethodByName(method)
		if methodRV.IsZero() {
			return runtime.MakeRet(nil, ErrMethodNotFound)
		}

		argsRV, err := prepareArgsRV(methodRV, args)
		if err != nil {
			return runtime.MakeRet(nil, err)
		}

		return runtime.MakeRet(methodRV.Call(argsRV), nil)

	}, plugin, method, args), nil
}

func (d *DistributedDispatcher) callEntity(entityId uid.Id, component, method string, args variant.Array) (asyncRet runtime.AsyncRet, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", golaxy.ErrPanicked, panicErr)
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
		if methodRV.IsZero() {
			return runtime.MakeRet(nil, ErrMethodNotFound)
		}

		argsRV, err := prepareArgsRV(methodRV, args)
		if err != nil {
			return runtime.MakeRet(nil, err)
		}

		return runtime.MakeRet(methodRV.Call(argsRV), nil)

	}, component, method, args), nil
}

func (d *DistributedDispatcher) reply(src string, corrId int64, retsRV []reflect.Value, retErr error) {
	msg := &gap.MsgRPCReply{
		CorrId: corrId,
	}

	var err error
	msg.Rets, err = variant.MakeArray(retsRV)
	if err != nil {
		log.Errorf(d.servCtx, "%d: reply to %q failed, %s", corrId, src, err)
		return
	}

	if retErr != nil {
		msg.Error = *variant.MakeError(retErr)
	}

	err = d.dist.SendMsg(src, msg)
	if err != nil {
		log.Errorf(d.servCtx, "%d: reply to %q failed, %s", corrId, src, err)
		return
	}

	log.Debugf(d.servCtx, "%d: reply to %q ok", corrId, src)
}

func (d *DistributedDispatcher) resolve(reply *gap.MsgRPCReply) error {
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
