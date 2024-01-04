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
	ctx     service.Context
	dist    distributed.Distributed
	watcher distributed.Watcher
}

// Init 初始化
func (d *DistributedDispatcher) Init(ctx service.Context) {
	d.ctx = ctx
	d.dist = distributed.Using(ctx)
	d.watcher = d.dist.WatchMsg(context.Background(), generic.CastDelegateFunc2(d.handleMsg))

	log.Debugf(d.ctx, "dispatcher %q started", types.AnyFullName(*d))
}

// Shut 结束
func (d *DistributedDispatcher) Shut(ctx service.Context) {
	<-d.watcher.Stop()

	log.Debugf(d.ctx, "dispatcher %q stopped", types.AnyFullName(*d))
}

func (d *DistributedDispatcher) handleMsg(src string, mp gap.MsgPacket) error {
	addr := d.dist.GetAddress()

	if !strings.HasPrefix(mp.Head.Src, addr.Domain) {
		return nil
	}

	switch mp.Head.MsgId {
	case gap.MsgId_OneWayRPC:
		msg := mp.Msg.(*gap.MsgOneWayRPC)
		go d.acceptNotify(msg)

	case gap.MsgId_RPC_Request:
		msg := mp.Msg.(*gap.MsgRPCRequest)
		go d.acceptRequest(mp.Head.Src, msg)

	case gap.MsgId_RPC_Reply:
		msg := mp.Msg.(*gap.MsgRPCReply)
		return d.resolve(msg)
	}

	return nil
}

func (d *DistributedDispatcher) acceptNotify(req *gap.MsgOneWayRPC) {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		log.Errorf(d.ctx, "parse call path %q failed, %s", req.Path, err)
		return
	}

	switch cp.Category {
	case callpath.Service:
		if _, err := d.callService(cp.Plugin, cp.Method, req.Args); err != nil {
			log.Errorf(d.ctx, "service plugin %q method %q calls failed, %s", cp.Plugin, cp.Method, err)
			return
		}

		log.Debugf(d.ctx, "service plugin %q method %q calls finished", cp.Plugin, cp.Method)
		return

	case callpath.Runtime:
		if _, err := d.callRuntime(uid.Id(cp.EntityId), cp.Plugin, cp.Method, req.Args); err != nil {
			log.Errorf(d.ctx, "entity %q runtime plugin %q method %q calls failed, %s", cp.EntityId, cp.Plugin, cp.Method, err)
			return
		}

		log.Debugf(d.ctx, "entity %q runtime plugin %q method %q calls finished", cp.EntityId, cp.Plugin, cp.Method)
		return

	case callpath.Entity:
		if _, err := d.callEntity(uid.Id(cp.EntityId), cp.Component, cp.Method, req.Args); err != nil {
			log.Errorf(d.ctx, "entity %q component %q method %q calls failed, %s", cp.EntityId, cp.Component, cp.Method, err)
			return
		}

		log.Debugf(d.ctx, "entity %q component %q method %q calls finished", cp.EntityId, cp.Component, cp.Method)
		return
	}
}

func (d *DistributedDispatcher) acceptRequest(src string, req *gap.MsgRPCRequest) {
	cp, err := callpath.Parse(req.Path)
	if err != nil {
		log.Errorf(d.ctx, "parse call path %q failed, %s", req.Path, err)
		d.reply(src, req.CorrId, nil, err)
		return
	}

	switch cp.Category {
	case callpath.Service:
		retsRV, err := d.callService(cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.ctx, "service plugin %q method %q calls failed, %s", cp.Plugin, cp.Method, err)
			d.reply(src, req.CorrId, nil, err)
			return
		}

		log.Debugf(d.ctx, "service plugin %q method %q calls finished", cp.Plugin, cp.Method)
		d.reply(src, req.CorrId, retsRV, nil)
		return

	case callpath.Runtime:
		retsRV, err := d.callRuntime(uid.Id(cp.EntityId), cp.Plugin, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.ctx, "entity %q runtime plugin %q method %q calls failed, %s", cp.EntityId, cp.Plugin, cp.Method, err)
			d.reply(src, req.CorrId, nil, err)
			return
		}

		log.Debugf(d.ctx, "entity %q runtime plugin %q method %q calls finished", cp.EntityId, cp.Plugin, cp.Method)
		d.reply(src, req.CorrId, retsRV, nil)
		return

	case callpath.Entity:
		retsRV, err := d.callEntity(uid.Id(cp.EntityId), cp.Component, cp.Method, req.Args)
		if err != nil {
			log.Errorf(d.ctx, "entity %q component %q method %q calls failed, %s", cp.EntityId, cp.Component, cp.Method, err)
			d.reply(src, req.CorrId, nil, err)
			return
		}

		log.Debugf(d.ctx, "entity %q component %q method %q calls finished", cp.EntityId, cp.Component, cp.Method)
		d.reply(src, req.CorrId, retsRV, nil)
		return
	}
}

func (d *DistributedDispatcher) callService(plugin, method string, args variant.Array) (rets []reflect.Value, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", golaxy.ErrPanicked, panicErr)
		}
	}()

	pi, ok := d.ctx.GetPluginBundle().Get(plugin)
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

func (d *DistributedDispatcher) callRuntime(entityId uid.Id, plugin, method string, args variant.Array) (rets []reflect.Value, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", golaxy.ErrPanicked, panicErr)
		}
	}()

	ret := d.ctx.Call(entityId, func(entity ec.Entity, a ...any) service.Ret {
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

	}, plugin, method, args).Wait(d.ctx)
	if !ret.OK() {
		return nil, ret.Error
	}

	return ret.Value.([]reflect.Value), nil
}

func (d *DistributedDispatcher) callEntity(entityId uid.Id, component, method string, args variant.Array) (rets []reflect.Value, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			err = fmt.Errorf("%w: %w", golaxy.ErrPanicked, panicErr)
		}
	}()

	ret := d.ctx.Call(entityId, func(entity ec.Entity, a ...any) service.Ret {
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

	}, component, method, args).Wait(d.ctx)
	if !ret.OK() {
		return nil, ret.Error
	}

	return ret.Value.([]reflect.Value), nil
}

func (d *DistributedDispatcher) reply(src string, corrId int64, retsRV []reflect.Value, retErr error) {
	msg := &gap.MsgRPCReply{
		CorrId: corrId,
	}

	var err error
	msg.Rets, err = variant.MakeArray(retsRV)
	if err != nil {
		log.Errorf(d.ctx, "reply to %q failed, corr_id:%d, %s", src, corrId, err)
		return
	}

	if retErr != nil {
		msg.Error = *variant.MakeError(retErr)
	}

	err = d.dist.SendMsg(src, msg)
	if err != nil {
		log.Errorf(d.ctx, "reply to %q failed, corr_id:%d, %s", src, corrId, err)
		return
	}

	log.Debugf(d.ctx, "reply to %q ok, corr_id:%d", src, corrId)
}

func (d *DistributedDispatcher) resolve(reply *gap.MsgRPCReply) error {
	rets := make([]any, 0, len(reply.Rets))
	for i := range reply.Rets {
		rets = append(rets, reply.Rets[i].Value)
	}

	var err error
	if !reply.Error.OK() {
		err = &reply.Error
	}

	return d.dist.GetFutures().Resolve(reply.CorrId, concurrent.MakeRet[any](rets, err))
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

		if !argRT.AssignableTo(paramRT) {
			if argRT.Elem().AssignableTo(paramRT) {
				argRV = argRV.Elem()
			} else {
				return nil, ErrMethodParameterTypeMismatch
			}
		}

		argsRV = append(argsRV, argRV)
	}

	return argsRV, nil
}
