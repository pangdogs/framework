package rpc

import (
	"errors"
	"kit.golaxy.org/golaxy/ec"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util/uid"
	"kit.golaxy.org/plugins/distributed"
	"kit.golaxy.org/plugins/gap"
	"kit.golaxy.org/plugins/gap/variant"
	"kit.golaxy.org/plugins/rpc/callpath"
	"kit.golaxy.org/plugins/util/concurrent"
	"reflect"
	"strings"
)

type DistributedDispatcher struct {
}

// Match 是否匹配
func (DistributedDispatcher) Match(ctx service.Context, src string) bool {
	addr := distributed.Using(ctx).GetAddress()

	if !strings.HasPrefix(src, addr.Domain) {
		return false
	}

	return true
}

// Dispatch 分发消息
func (DistributedDispatcher) Dispatch(ctx service.Context, src string, msg gap.Msg) error {
	dist := distributed.Using(ctx)

	switch m := msg.(type) {
	case *gap.MsgOneWayRPC:
		cp, err := callpath.Parse(m.Path)
		if err != nil {
			return err
		}

		switch cp.Category {
		case callpath.Service:
			go func() {
				pi, ok := ctx.GetPluginBundle().Get(cp.Plugin)
				if !ok {
					return
				}

				methodRV := pi.Reflected.MethodByName(cp.Method)
				if methodRV.IsZero() {
					return
				}

				argsRV := make([]reflect.Value, 0, len(m.Args))
				for i := range m.Args {
					argsRV = append(argsRV, m.Args[i].Reflected)
				}

				methodRV.Call(argsRV)
			}()

		case callpath.Runtime:
			ctx.Call(uid.Id(cp.EntityId), func(entity ec.Entity, a ...any) service.Ret {

			}, cp.Plugin, cp.Method, m.Args)

		case callpath.Entity:
			ctx.CallVoid(uid.Id(cp.EntityId), func(entity ec.Entity, a ...any) {
				compName := a[0].(string)
				method := a[1].(string)
				args := a[2].([]variant.Array)

				comp := entity.GetComponent(compName)
				if comp == nil {
					return
				}

				methodRV := ec.UnsafeComponent(comp).GetReflected().MethodByName(method)
				if methodRV.IsZero() {
					return
				}

				argsRV := make([]reflect.Value, 0, len(args))
				for i := range args {
					argsRV = append(argsRV, m.Args[i].Reflected)
				}

				methodRV.Call(argsRV)

			}, cp.Component, cp.Method, m.Args)
		}

	case *gap.MsgRPCRequest:

	case *gap.MsgRPCReply:
		argsRV := make([]reflect.Value, 0, len(args))
		for i := range args {
			argsRV = append(argsRV, m.Args[i].Reflected)
		}

		return dist.GetFutures().Resolve(m.CorrId, concurrent.MakeRet[any]())

	default:
		return errors.New("invalid msg type")
	}
}
