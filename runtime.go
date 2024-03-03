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
	etcd_client "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"sync"
)

type _IRuntime interface {
	setup(servCtx service.Context, composite any)
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

// RuntimeBehavior 运行时行为，开发新运行环境时，匿名嵌入至新运行时结构体中
type RuntimeBehavior struct {
	servCtx   service.Context
	composite any
}

func (rb *RuntimeBehavior) setup(servCtx service.Context, composite any) {
	rb.servCtx = servCtx
	rb.composite = composite
}

func (rb *RuntimeBehavior) generate(settings _RuntimeSettings) core.Runtime {
	startupConf := rb.GetStartupConf()

	face := iface.Face[runtime.Context]{}

	if cb, ok := rb.composite.(SetupRuntimeContextComposite); ok {
		face = iface.MakeFace(cb.MakeContextComposite())
	}
	ctxFrameLoopBeginCB, _ := face.Iface.(LifecycleRuntimeContextFrameLoopBegin)
	ctxFrameUpdateBeginCB, _ := face.Iface.(LifecycleRuntimeContextFrameUpdateBegin)
	ctxFrameUpdateEndCB, _ := face.Iface.(LifecycleRuntimeContextFrameUpdateEnd)
	ctxFrameLoopEndCB, _ := face.Iface.(LifecycleRuntimeContextFrameLoopEnd)
	ctxRunCallBeginCB, _ := face.Iface.(LifecycleRuntimeContextRunCallBegin)
	ctxRunCallEndCB, _ := face.Iface.(LifecycleRuntimeContextRunCallEnd)
	ctxRunGCBeginCB, _ := face.Iface.(LifecycleRuntimeContextRunGCBegin)
	ctxRunGCEndCB, _ := face.Iface.(LifecycleRuntimeContextRunGCEnd)

	frameLoopBeginCB, _ := rb.composite.(LifecycleRuntimeFrameLoopBegin)
	frameUpdateBeginCB, _ := rb.composite.(LifecycleRuntimeFrameUpdateBegin)
	frameUpdateEndCB, _ := rb.composite.(LifecycleRuntimeFrameUpdateEnd)
	frameLoopEndCB, _ := rb.composite.(LifecycleRuntimeFrameLoopEnd)
	runCallBeginCB, _ := rb.composite.(LifecycleRuntimeRunCallBegin)
	runCallEndCB, _ := rb.composite.(LifecycleRuntimeRunCallEnd)
	runGCBeginCB, _ := rb.composite.(LifecycleRuntimeRunGCBegin)
	runGCEndCB, _ := rb.composite.(LifecycleRuntimeRunGCEnd)

	rtCtx := runtime.NewContext(rb.GetServiceCtx(),
		runtime.With.Context.CompositeFace(face),
		runtime.With.Context.Name(settings.Name),
		runtime.With.Context.PanicHandling(settings.AutoRecover, settings.ReportError),
		runtime.With.Context.RunningHandler(generic.CastDelegateAction2(func(ctx runtime.Context, state runtime.RunningState) {
			switch state {
			case runtime.RunningState_Birth:
				if cb, ok := rb.composite.(LifecycleRuntimeBirth); ok {
					cb.Birth(ctx)
				}
				if cb, ok := ctx.(LifecycleRuntimeContextBirth); ok {
					cb.Birth()
				}
			case runtime.RunningState_Starting:
				if cb, ok := rb.composite.(LifecycleRuntimeStarting); ok {
					cb.Starting(ctx)
				}
				if cb, ok := ctx.(LifecycleRuntimeContextStarting); ok {
					cb.Starting()
				}
			case runtime.RunningState_Started:
				if cb, ok := rb.composite.(LifecycleRuntimeStarted); ok {
					cb.Started(ctx)
				}
				if cb, ok := ctx.(LifecycleRuntimeContextStarted); ok {
					cb.Started()
				}
			case runtime.RunningState_FrameLoopBegin:
				if cb := frameLoopBeginCB; cb != nil {
					cb.FrameLoopBegin(ctx)
				}
				if cb := ctxFrameLoopBeginCB; cb != nil {
					cb.FrameLoopBegin()
				}
			case runtime.RunningState_FrameUpdateBegin:
				if cb := frameUpdateBeginCB; cb != nil {
					cb.FrameUpdateBegin(ctx)
				}
				if cb := ctxFrameUpdateBeginCB; cb != nil {
					cb.FrameUpdateBegin()
				}
			case runtime.RunningState_FrameUpdateEnd:
				if cb := frameUpdateEndCB; cb != nil {
					cb.FrameUpdateEnd(ctx)
				}
				if cb := ctxFrameUpdateEndCB; cb != nil {
					cb.FrameUpdateEnd()
				}
			case runtime.RunningState_FrameLoopEnd:
				if cb := frameLoopEndCB; cb != nil {
					cb.FrameLoopEnd(ctx)
				}
				if cb := ctxFrameLoopEndCB; cb != nil {
					cb.FrameLoopEnd()
				}
			case runtime.RunningState_RunCallBegin:
				if cb := runCallBeginCB; cb != nil {
					cb.RunCallBegin(ctx)
				}
				if cb := ctxRunCallBeginCB; cb != nil {
					cb.RunCallBegin()
				}
			case runtime.RunningState_RunCallEnd:
				if cb := runCallEndCB; cb != nil {
					cb.RunCallEnd(ctx)
				}
				if cb := ctxRunCallEndCB; cb != nil {
					cb.RunCallEnd()
				}
			case runtime.RunningState_RunGCBegin:
				if cb := runGCBeginCB; cb != nil {
					cb.RunGCBegin(ctx)
				}
				if cb := ctxRunGCBeginCB; cb != nil {
					cb.RunGCBegin()
				}
			case runtime.RunningState_RunGCEnd:
				if cb := runGCEndCB; cb != nil {
					cb.RunGCEnd(ctx)
				}
				if cb := ctxRunGCEndCB; cb != nil {
					cb.RunGCEnd()
				}
			case runtime.RunningState_Terminating:
				if cb, ok := rb.composite.(LifecycleRuntimeTerminating); ok {
					cb.Terminating(ctx)
				}
				if cb, ok := ctx.(LifecycleRuntimeContextTerminating); ok {
					cb.Terminating()
				}
			case runtime.RunningState_Terminated:
				if cb, ok := rb.composite.(LifecycleRuntimeTerminated); ok {
					cb.Terminated(ctx)
				}
				if cb, ok := ctx.(LifecycleRuntimeContextTerminated); ok {
					cb.Terminated()
				}
			}
		})),
	)

	// 安装日志插件
	if cb, ok := rb.composite.(InstallRuntimeLogger); ok {
		cb.InstallLogger(rtCtx)
	}
	if _, ok := rtCtx.GetPluginBundle().Get(log.Name); !ok {
		if v, _ := rb.GetMemKVs().Load("zap.logger"); v != nil {
			zap_log.Install(rtCtx,
				zap_log.With.ZapLogger(v.(*zap.Logger)),
				zap_log.With.ServiceInfo(true),
				zap_log.With.RuntimeInfo(true),
			)
		}
	}

	// 安装分布式实体支持插件
	if cb, ok := rb.composite.(InstallRuntimeDistEntities); ok {
		cb.InstallDistEntities(rtCtx)
	}
	if _, ok := rtCtx.GetPluginBundle().Get(dent.Name); !ok {
		v, _ := rb.GetMemKVs().Load("etcd.init_client")
		fun, _ := v.(func())
		if fun == nil {
			panic("service memory etcd.init_client not existed")
		}
		fun()

		v, _ = rb.GetMemKVs().Load("etcd.client")
		cli, _ := v.(*etcd_client.Client)
		if cli == nil {
			panic("service memory etcd.client not existed")
		}

		dent.Install(rtCtx,
			dent.With.EtcdClient(cli),
			dent.With.TTL(startupConf.GetDuration("service.dent_ttl")),
		)
	}

	// 初始化回调
	if cb, ok := rb.composite.(LifecycleRuntimeInit); ok {
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
func (rb *RuntimeBehavior) GetServiceCtx() service.Context {
	return rb.servCtx
}

// GetStartupConf 获取启动参数配置
func (rb *RuntimeBehavior) GetStartupConf() *viper.Viper {
	v, _ := rb.GetMemKVs().Load("startup.conf")
	if v == nil {
		panic("service memory startup.conf not existed")
	}
	return v.(*viper.Viper)
}

// GetMemKVs 获取服务内存KV数据库
func (rb *RuntimeBehavior) GetMemKVs() *sync.Map {
	memKVs, _ := rb.GetServiceCtx().Value("mem_kvs").(*sync.Map)
	if memKVs == nil {
		panic("service memory not existed")
	}
	return memKVs
}
