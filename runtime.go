package framework

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/iface"
	"git.golaxy.org/core/utils/reinterpret"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/plugins/conf"
	"git.golaxy.org/framework/plugins/dentr"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/log/zap_log"
	"git.golaxy.org/framework/plugins/rpcstack"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"reflect"
)

type iRuntimeGeneric interface {
	init(ctx service.Context, composite any)
	generate(settings _RuntimeSettings) core.Runtime
}

type _RuntimeSettings struct {
	Name                 string
	PersistId            uid.Id
	AutoRecover          bool
	ReportError          chan error
	FPS                  float32
	ProcessQueueCapacity int
}

// RuntimeGenericT 运行时泛化类型实例化
type RuntimeGenericT[T any] struct {
	RuntimeGeneric
}

// Instantiation 实例化
func (r *RuntimeGenericT[T]) Instantiation() IRuntimeInstance {
	return reflect.New(reflect.TypeFor[T]()).Interface().(IRuntimeInstance)
}

// RuntimeGeneric 运行时泛化类型
type RuntimeGeneric struct {
	serv      IServiceInstance
	composite any
}

func (r *RuntimeGeneric) init(ctx service.Context, composite any) {
	r.serv = reinterpret.Cast[IServiceInstance](ctx)
	r.composite = composite
}

func (r *RuntimeGeneric) generate(settings _RuntimeSettings) core.Runtime {
	wholeConf := conf.Using(r.serv).Whole()

	face := iface.Face[runtime.Context]{}

	if cb, ok := r.composite.(IRuntimeInstantiation); ok {
		face = iface.MakeFaceTReflectC[runtime.Context, IRuntimeInstance](cb.Instantiation())
	}

	iFrameLoopBeginCB, _ := face.Iface.(LifecycleRuntimeFrameLoopBegin)
	iFrameUpdateBeginCB, _ := face.Iface.(LifecycleRuntimeFrameUpdateBegin)
	iFrameUpdateEndCB, _ := face.Iface.(LifecycleRuntimeFrameUpdateEnd)
	iFrameLoopEndCB, _ := face.Iface.(LifecycleRuntimeFrameLoopEnd)
	iRunCallBeginCB, _ := face.Iface.(LifecycleRuntimeRunCallBegin)
	iRunCallEndCB, _ := face.Iface.(LifecycleRuntimeRunCallEnd)
	iRunGCBeginCB, _ := face.Iface.(LifecycleRuntimeRunGCBegin)
	iRunGCEndCB, _ := face.Iface.(LifecycleRuntimeRunGCEnd)

	frameLoopBeginCB, _ := r.composite.(LifecycleRuntimeFrameLoopBegin)
	frameUpdateBeginCB, _ := r.composite.(LifecycleRuntimeFrameUpdateBegin)
	frameUpdateEndCB, _ := r.composite.(LifecycleRuntimeFrameUpdateEnd)
	frameLoopEndCB, _ := r.composite.(LifecycleRuntimeFrameLoopEnd)
	runCallBeginCB, _ := r.composite.(LifecycleRuntimeRunCallBegin)
	runCallEndCB, _ := r.composite.(LifecycleRuntimeRunCallEnd)
	runGCBeginCB, _ := r.composite.(LifecycleRuntimeRunGCBegin)
	runGCEndCB, _ := r.composite.(LifecycleRuntimeRunGCEnd)

	rtCtx := runtime.NewContext(r.GetService(),
		runtime.With.Context.CompositeFace(face),
		runtime.With.Context.Name(settings.Name),
		runtime.With.Context.PersistId(settings.PersistId),
		runtime.With.Context.PanicHandling(settings.AutoRecover, settings.ReportError),
		runtime.With.Context.RunningHandler(generic.MakeDelegateAction2(func(ctx runtime.Context, state runtime.RunningState) {
			inst := reinterpret.Cast[IRuntimeInstance](ctx)

			switch state {
			case runtime.RunningState_Birth:
				if cb, ok := r.composite.(LifecycleRuntimeBirth); ok {
					cb.Birth(inst)
				}
				if cb, ok := inst.(LifecycleRuntimeBirth); ok {
					cb.Birth(inst)
				}
			case runtime.RunningState_Starting:
				if cb, ok := r.composite.(LifecycleRuntimeStarting); ok {
					cb.Starting(inst)
				}
				if cb, ok := inst.(LifecycleRuntimeStarting); ok {
					cb.Starting(inst)
				}
			case runtime.RunningState_Started:
				if cb, ok := r.composite.(LifecycleRuntimeStarted); ok {
					cb.Started(inst)
				}
				if cb, ok := inst.(LifecycleRuntimeStarted); ok {
					cb.Started(inst)
				}
			case runtime.RunningState_FrameLoopBegin:
				if cb := frameLoopBeginCB; cb != nil {
					cb.FrameLoopBegin(inst)
				}
				if cb := iFrameLoopBeginCB; cb != nil {
					cb.FrameLoopBegin(inst)
				}
			case runtime.RunningState_FrameUpdateBegin:
				if cb := frameUpdateBeginCB; cb != nil {
					cb.FrameUpdateBegin(inst)
				}
				if cb := iFrameUpdateBeginCB; cb != nil {
					cb.FrameUpdateBegin(inst)
				}
			case runtime.RunningState_FrameUpdateEnd:
				if cb := frameUpdateEndCB; cb != nil {
					cb.FrameUpdateEnd(inst)
				}
				if cb := iFrameUpdateEndCB; cb != nil {
					cb.FrameUpdateEnd(inst)
				}
			case runtime.RunningState_FrameLoopEnd:
				if cb := frameLoopEndCB; cb != nil {
					cb.FrameLoopEnd(inst)
				}
				if cb := iFrameLoopEndCB; cb != nil {
					cb.FrameLoopEnd(inst)
				}
			case runtime.RunningState_RunCallBegin:
				if cb := runCallBeginCB; cb != nil {
					cb.RunCallBegin(inst)
				}
				if cb := iRunCallBeginCB; cb != nil {
					cb.RunCallBegin(inst)
				}
			case runtime.RunningState_RunCallEnd:
				if cb := runCallEndCB; cb != nil {
					cb.RunCallEnd(inst)
				}
				if cb := iRunCallEndCB; cb != nil {
					cb.RunCallEnd(inst)
				}
			case runtime.RunningState_RunGCBegin:
				if cb := runGCBeginCB; cb != nil {
					cb.RunGCBegin(inst)
				}
				if cb := iRunGCBeginCB; cb != nil {
					cb.RunGCBegin(inst)
				}
			case runtime.RunningState_RunGCEnd:
				if cb := runGCEndCB; cb != nil {
					cb.RunGCEnd(inst)
				}
				if cb := iRunGCEndCB; cb != nil {
					cb.RunGCEnd(inst)
				}
			case runtime.RunningState_Terminating:
				if cb, ok := r.composite.(LifecycleRuntimeTerminating); ok {
					cb.Terminating(inst)
				}
				if cb, ok := inst.(LifecycleRuntimeTerminating); ok {
					cb.Terminating(inst)
				}
			case runtime.RunningState_Terminated:
				if cb, ok := r.composite.(LifecycleRuntimeTerminated); ok {
					cb.Terminated(inst)
				}
				if cb, ok := inst.(LifecycleRuntimeTerminated); ok {
					cb.Terminated(inst)
				}
			}
		})),
	)

	rtInst := reinterpret.Cast[IRuntimeInstance](rtCtx)

	installed := func(name string) bool {
		_, ok := rtInst.GetPluginBundle().Get(name)
		return ok
	}

	// 安装日志插件
	if !installed(log.Name) {
		if cb, ok := rtInst.(InstallRuntimeLogger); ok {
			cb.InstallLogger(rtInst)
		}
	}
	if !installed(log.Name) {
		if cb, ok := r.composite.(InstallRuntimeLogger); ok {
			cb.InstallLogger(rtInst)
		}
	}
	if !installed(log.Name) {
		if v, _ := r.serv.GetMemKV().Load("zap.logger"); v != nil {
			zap_log.Install(rtInst,
				zap_log.With.ZapLogger(v.(*zap.Logger)),
				zap_log.With.ServiceInfo(wholeConf.GetBool("log.service_info")),
				zap_log.With.RuntimeInfo(wholeConf.GetBool("log.runtime_info")),
			)
		}
	}

	// 安装RPC调用堆栈支持
	if !installed(rpcstack.Name) {
		if cb, ok := rtInst.(InstallRuntimeRPCStack); ok {
			cb.InstallRPCStack(rtInst)
		}
	}
	if !installed(rpcstack.Name) {
		if cb, ok := r.composite.(InstallRuntimeRPCStack); ok {
			cb.InstallRPCStack(rtInst)
		}
	}
	if !installed(rpcstack.Name) {
		rpcstack.Install(rtInst)
	}

	// 安装分布式实体支持插件
	if !installed(dentr.Name) {
		if cb, ok := rtInst.(InstallRuntimeDistEntityRegistry); ok {
			cb.InstallDistEntityRegistry(rtInst)
		}
	}
	if !installed(dentr.Name) {
		if cb, ok := r.composite.(InstallRuntimeDistEntityRegistry); ok {
			cb.InstallDistEntityRegistry(rtInst)
		}
	}
	if !installed(dentr.Name) {
		v, _ := r.GetService().GetMemKV().Load("etcd.init_client")
		fun, _ := v.(func())
		if fun == nil {
			panic("service memory kv etcd.init_client not existed")
		}
		fun()

		v, _ = r.GetService().GetMemKV().Load("etcd.client")
		cli, _ := v.(*etcdv3.Client)
		if cli == nil {
			panic("service memory kv etcd.client not existed")
		}

		dentr.Install(rtInst,
			dentr.With.EtcdClient(cli),
			dentr.With.TTL(wholeConf.GetDuration("service.dent_ttl")),
		)
	}

	// 组装完成回调回调
	if cb, ok := r.composite.(LifecycleRuntimeBuilt); ok {
		cb.Built(rtInst)
	}
	if cb, ok := rtInst.(LifecycleRuntimeBuilt); ok {
		cb.Built(rtInst)
	}

	return core.NewRuntime(rtInst,
		core.With.Runtime.Frame(func() runtime.Frame {
			if settings.FPS <= 0 {
				return nil
			}
			return runtime.NewFrame(
				runtime.With.Frame.TargetFPS(settings.FPS),
			)
		}()),
		core.With.Runtime.AutoRun(true),
		core.With.Runtime.ProcessQueueCapacity(settings.ProcessQueueCapacity),
	)
}

// GetService 获取服务
func (r *RuntimeGeneric) GetService() IServiceInstance {
	return r.serv
}
