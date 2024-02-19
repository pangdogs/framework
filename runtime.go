package framework

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/framework/plugins/dent"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/plugins/log/zap_log"
	"github.com/spf13/viper"
	etcd_client "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"sync"
)

type _IRuntime interface {
	init(servCtx service.Context, composite any)
	generate(settings _RuntimeSettings) core.Runtime
}

type _RuntimeSettings struct {
	Name                 string
	AutoRecover          bool
	ReportError          chan error
	FrameFPS             float32
	FrameBlink           bool
	AutoRun              bool
	ProcessQueueCapacity int
}

// RuntimeBehavior 运行时行为，开发新运行环境时，匿名嵌入至新运行时结构体中
type RuntimeBehavior struct {
	servCtx   service.Context
	composite any
}

func (rb *RuntimeBehavior) init(servCtx service.Context, composite any) {
	rb.servCtx = servCtx
	rb.composite = composite
}

func (rb *RuntimeBehavior) generate(settings _RuntimeSettings) core.Runtime {
	startupConf := rb.GetStartupConf()

	rtCtx := runtime.NewContext(rb.GetServiceCtx(),
		runtime.Option{}.Context.Name(settings.Name),
		runtime.Option{}.Context.AutoRecover(settings.AutoRecover),
		runtime.Option{}.Context.ReportError(settings.ReportError),
		runtime.Option{}.Context.RunningHandler(generic.CastDelegateAction2(func(ctx runtime.Context, state runtime.RunningState) {
			switch state {
			case runtime.RunningState_Birth:
				if cb, ok := rb.composite.(LifecycleRuntimeBirth); ok {
					cb.Birth(ctx)
				}
			case runtime.RunningState_Starting:
				if cb, ok := rb.composite.(LifecycleRuntimeStarting); ok {
					cb.Starting(ctx)
				}
			case runtime.RunningState_Started:
				if cb, ok := rb.composite.(LifecycleRuntimeStarted); ok {
					cb.Started(ctx)
				}
			case runtime.RunningState_FrameLoopBegin:
				if cb, ok := rb.composite.(LifecycleRuntimeFrameLoopBegin); ok {
					cb.FrameLoopBegin(ctx)
				}
			case runtime.RunningState_FrameUpdateEnd:
				if cb, ok := rb.composite.(LifecycleRuntimeFrameUpdateEnd); ok {
					cb.FrameUpdateEnd(ctx)
				}
			case runtime.RunningState_FrameUpdateBegin:
				if cb, ok := rb.composite.(LifecycleRuntimeFrameUpdateBegin); ok {
					cb.FrameUpdateBegin(ctx)
				}
			case runtime.RunningState_FrameLoopEnd:
				if cb, ok := rb.composite.(LifecycleRuntimeFrameLoopEnd); ok {
					cb.FrameLoopEnd(ctx)
				}
			case runtime.RunningState_RunCallBegin:
				if cb, ok := rb.composite.(LifecycleRuntimeRunCallBegin); ok {
					cb.RunCallBegin(ctx)
				}
			case runtime.RunningState_RunCallEnd:
				if cb, ok := rb.composite.(LifecycleRuntimeRunCallEnd); ok {
					cb.RunCallEnd(ctx)
				}
			case runtime.RunningState_RunGCBegin:
				if cb, ok := rb.composite.(LifecycleRuntimeRunGCBegin); ok {
					cb.RunGCBegin(ctx)
				}
			case runtime.RunningState_RunGCEnd:
				if cb, ok := rb.composite.(LifecycleRuntimeRunGCEnd); ok {
					cb.RunGCEnd(ctx)
				}
			case runtime.RunningState_Terminating:
				if cb, ok := rb.composite.(LifecycleRuntimeTerminating); ok {
					cb.Terminating(ctx)
				}
			case runtime.RunningState_Terminated:
				if cb, ok := rb.composite.(LifecycleRuntimeTerminated); ok {
					cb.Terminated(ctx)
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
				zap_log.Option{}.ZapLogger(v.(*zap.Logger)),
				zap_log.Option{}.ServiceInfo(true),
				zap_log.Option{}.RuntimeInfo(true),
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
			dent.Option{}.EtcdClient(cli),
			dent.Option{}.TTL(startupConf.GetDuration("service.dent_ttl")),
		)
	}

	// 初始化回调
	if cb, ok := rb.composite.(LifecycleRuntimeInit); ok {
		cb.Init(rtCtx)
	}

	return core.NewRuntime(rtCtx,
		core.Option{}.Runtime.Frame(func() runtime.Frame {
			if settings.FrameFPS <= 0 {
				return nil
			}
			return runtime.NewFrame(
				runtime.Option{}.Frame.TargetFPS(settings.FrameFPS),
				runtime.Option{}.Frame.Blink(settings.FrameBlink),
			)
		}()),
		core.Option{}.Runtime.AutoRun(settings.AutoRun),
		core.Option{}.Runtime.ProcessQueueCapacity(settings.ProcessQueueCapacity),
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
