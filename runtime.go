package framework

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/iface"
	"git.golaxy.org/framework/plugins/dent"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/log/zap_log"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"sync"
)

type _IRuntimeGeneric interface {
	setup(ctx service.Context, composite any)
	generate(settings _RuntimeSettings) core.Runtime
}

type _RuntimeSettings struct {
	Name                 string
	AutoRecover          bool
	ReportError          chan error
	FPS                  float32
	AutoRun              bool
	ProcessQueueCapacity int
}

// RuntimeGeneric 运行时泛化类型
type RuntimeGeneric struct {
	servCtx   service.Context
	composite any
}

func (r *RuntimeGeneric) setup(ctx service.Context, composite any) {
	r.servCtx = ctx
	r.composite = composite
}

func (r *RuntimeGeneric) generate(settings _RuntimeSettings) core.Runtime {
	startupConf := r.GetStartupConf()

	face := iface.Face[runtime.Context]{}

	if cb, ok := r.composite.(IRuntimeInstantiation); ok {
		face = iface.MakeFace(cb.Instantiation())
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

	rtCtx := runtime.NewContext(r.GetServiceCtx(),
		runtime.With.Context.CompositeFace(face),
		runtime.With.Context.Name(settings.Name),
		runtime.With.Context.PanicHandling(settings.AutoRecover, settings.ReportError),
		runtime.With.Context.RunningHandler(generic.MakeDelegateAction2(func(ctx runtime.Context, state runtime.RunningState) {
			switch state {
			case runtime.RunningState_Birth:
				if cb, ok := r.composite.(LifecycleRuntimeBirth); ok {
					cb.Birth(ctx)
				}
				if cb, ok := ctx.(LifecycleRuntimeBirth); ok {
					cb.Birth(ctx)
				}
			case runtime.RunningState_Starting:
				if cb, ok := r.composite.(LifecycleRuntimeStarting); ok {
					cb.Starting(ctx)
				}
				if cb, ok := ctx.(LifecycleRuntimeStarting); ok {
					cb.Starting(ctx)
				}
			case runtime.RunningState_Started:
				if cb, ok := r.composite.(LifecycleRuntimeStarted); ok {
					cb.Started(ctx)
				}
				if cb, ok := ctx.(LifecycleRuntimeStarted); ok {
					cb.Started(ctx)
				}
			case runtime.RunningState_FrameLoopBegin:
				if cb := frameLoopBeginCB; cb != nil {
					cb.FrameLoopBegin(ctx)
				}
				if cb := iFrameLoopBeginCB; cb != nil {
					cb.FrameLoopBegin(ctx)
				}
			case runtime.RunningState_FrameUpdateBegin:
				if cb := frameUpdateBeginCB; cb != nil {
					cb.FrameUpdateBegin(ctx)
				}
				if cb := iFrameUpdateBeginCB; cb != nil {
					cb.FrameUpdateBegin(ctx)
				}
			case runtime.RunningState_FrameUpdateEnd:
				if cb := frameUpdateEndCB; cb != nil {
					cb.FrameUpdateEnd(ctx)
				}
				if cb := iFrameUpdateEndCB; cb != nil {
					cb.FrameUpdateEnd(ctx)
				}
			case runtime.RunningState_FrameLoopEnd:
				if cb := frameLoopEndCB; cb != nil {
					cb.FrameLoopEnd(ctx)
				}
				if cb := iFrameLoopEndCB; cb != nil {
					cb.FrameLoopEnd(ctx)
				}
			case runtime.RunningState_RunCallBegin:
				if cb := runCallBeginCB; cb != nil {
					cb.RunCallBegin(ctx)
				}
				if cb := iRunCallBeginCB; cb != nil {
					cb.RunCallBegin(ctx)
				}
			case runtime.RunningState_RunCallEnd:
				if cb := runCallEndCB; cb != nil {
					cb.RunCallEnd(ctx)
				}
				if cb := iRunCallEndCB; cb != nil {
					cb.RunCallEnd(ctx)
				}
			case runtime.RunningState_RunGCBegin:
				if cb := runGCBeginCB; cb != nil {
					cb.RunGCBegin(ctx)
				}
				if cb := iRunGCBeginCB; cb != nil {
					cb.RunGCBegin(ctx)
				}
			case runtime.RunningState_RunGCEnd:
				if cb := runGCEndCB; cb != nil {
					cb.RunGCEnd(ctx)
				}
				if cb := iRunGCEndCB; cb != nil {
					cb.RunGCEnd(ctx)
				}
			case runtime.RunningState_Terminating:
				if cb, ok := r.composite.(LifecycleRuntimeTerminating); ok {
					cb.Terminating(ctx)
				}
				if cb, ok := ctx.(LifecycleRuntimeTerminating); ok {
					cb.Terminating(ctx)
				}
			case runtime.RunningState_Terminated:
				if cb, ok := r.composite.(LifecycleRuntimeTerminated); ok {
					cb.Terminated(ctx)
				}
				if cb, ok := ctx.(LifecycleRuntimeTerminated); ok {
					cb.Terminated(ctx)
				}
			}
		})),
	)

	// 安装日志插件
	if cb, ok := r.composite.(InstallRuntimeLogger); ok {
		cb.InstallLogger(rtCtx)
	}
	if _, ok := rtCtx.GetPluginBundle().Get(log.Name); !ok {
		if v, _ := r.GetMemKVs().Load("zap.logger"); v != nil {
			zap_log.Install(rtCtx,
				zap_log.With.ZapLogger(v.(*zap.Logger)),
				zap_log.With.ServiceInfo(startupConf.GetBool("log.service_info")),
				zap_log.With.RuntimeInfo(startupConf.GetBool("log.runtime_info")),
			)
		}
	}

	// 安装分布式实体支持插件
	if cb, ok := r.composite.(InstallRuntimeDistEntities); ok {
		cb.InstallDistEntities(rtCtx)
	}
	if _, ok := rtCtx.GetPluginBundle().Get(dent.Name); !ok {
		v, _ := r.GetMemKVs().Load("etcd.init_client")
		fun, _ := v.(func())
		if fun == nil {
			panic("service memory etcd.init_client not existed")
		}
		fun()

		v, _ = r.GetMemKVs().Load("etcd.client")
		cli, _ := v.(*etcdv3.Client)
		if cli == nil {
			panic("service memory etcd.client not existed")
		}

		dent.Install(rtCtx,
			dent.With.EtcdClient(cli),
			dent.With.TTL(startupConf.GetDuration("service.dent_ttl")),
		)
	}

	// 初始化回调
	if cb, ok := r.composite.(LifecycleRuntimeInit); ok {
		cb.Init(rtCtx)
	}

	return core.NewRuntime(rtCtx,
		core.With.Runtime.Frame(func() runtime.Frame {
			if settings.FPS <= 0 {
				return nil
			}
			return runtime.NewFrame(
				runtime.With.Frame.TargetFPS(settings.FPS),
			)
		}()),
		core.With.Runtime.AutoRun(settings.AutoRun),
		core.With.Runtime.ProcessQueueCapacity(settings.ProcessQueueCapacity),
	)
}

// GetServiceCtx 获取服务上下文
func (r *RuntimeGeneric) GetServiceCtx() service.Context {
	return r.servCtx
}

// GetStartupConf 获取启动参数配置
func (r *RuntimeGeneric) GetStartupConf() *viper.Viper {
	v, _ := r.GetMemKVs().Load("startup.conf")
	if v == nil {
		panic("service memory startup.conf not existed")
	}
	return v.(*viper.Viper)
}

// GetMemKVs 获取服务内存KV数据库
func (r *RuntimeGeneric) GetMemKVs() *sync.Map {
	memKVs, _ := r.GetServiceCtx().Value("mem_kvs").(*sync.Map)
	if memKVs == nil {
		panic("service memory not existed")
	}
	return memKVs
}
