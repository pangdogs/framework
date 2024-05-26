package framework

import (
	"context"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/generic"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// NewApp 创建应用
func NewApp() *App {
	return &App{}
}

type _ServPT struct {
	generic iServiceGeneric
	num     int
}

// App 应用
type App struct {
	servicePTs                       map[string]*_ServPT
	startupConf                      *viper.Viper
	initCB, startingCB, terminatedCB generic.DelegateAction1[*App]
}

func (app *App) lazyInit() {
	if app.servicePTs == nil {
		app.servicePTs = make(map[string]*_ServPT)
	}
	if app.startupConf == nil {
		app.startupConf = viper.New()
	}
}

// Setup 安装服务泛化类型
func (app *App) Setup(name string, generic any) *App {
	app.lazyInit()

	if generic == nil {
		panic(fmt.Errorf("%w: generic is nil", core.ErrArgs))
	}

	_generic, ok := generic.(iServiceGeneric)
	if !ok {
		panic(fmt.Errorf("%w: incorrect generic type", core.ErrArgs))
	}

	app.servicePTs[name] = &_ServPT{
		generic: _generic,
		num:     1,
	}

	return app
}

// InitCB 初始化回调
func (app *App) InitCB(cb generic.DelegateAction1[*App]) *App {
	app.initCB = cb
	return app
}

// StartingCB 启动回调
func (app *App) StartingCB(cb generic.DelegateAction1[*App]) *App {
	app.startingCB = cb
	return app
}

// TerminateCB 终止回调
func (app *App) TerminateCB(cb generic.DelegateAction1[*App]) *App {
	app.terminatedCB = cb
	return app
}

// Run 运行
func (app *App) Run() {
	app.lazyInit()

	// 日志参数
	pflag.String("log.format", "console", "logging output format (json|console)")
	pflag.String("log.level", "info", "logging level")
	pflag.String("log.dir", "./log/", "logging directory path")
	pflag.Int("log.size", 100*1024*1024, "log file splitting size")
	pflag.Bool("log.stdout", false, "logging output to stdout")
	pflag.Bool("log.development", false, "logging in development mode")
	pflag.Bool("log.service_info", false, "logging output service info")
	pflag.Bool("log.runtime_info", false, "logging output generic info")

	// 配置参数
	pflag.String("conf.format", "json", "config file format")
	pflag.String("conf.local_path", "", "local config file path")
	pflag.String("conf.remote_provider", "", "remote config provider")
	pflag.String("conf.remote_endpoint", "", "remote config endpoint")
	pflag.String("conf.remote_path", "", "remote config file path")
	pflag.Bool("conf.auto_update", true, "auto update config")

	// nats参数
	pflag.String("nats.address", "localhost:4222", "nats address")
	pflag.String("nats.username", "", "nats auth username")
	pflag.String("nats.password", "", "nats auth password")

	// etcd参数
	pflag.String("etcd.address", "localhost:2379", "etcd address")
	pflag.String("etcd.username", "", "etcd auth username")
	pflag.String("etcd.password", "", "etcd auth password")

	// 分布式服务参数
	pflag.String("service.version", "v0.0.0", "service version info")
	pflag.StringToString("service.meta", map[string]string{}, "service meta info")
	pflag.Duration("service.ttl", 10*time.Second, "ttl for service keepalive")
	pflag.Duration("service.future_timeout", 3*time.Second, "timeout for future model of service interaction")
	pflag.Duration("service.dent_ttl", 10*time.Second, "ttl for distributed entity keepalive")
	pflag.Bool("service.auto_recover", false, "enable panic auto recover")

	// 启动的服务列表
	pflag.StringToString("startup.services", func() map[string]string {
		ret := map[string]string{}
		for name, pt := range app.servicePTs {
			ret[name] = strconv.Itoa(pt.num)
		}
		return ret
	}(), "instances required for each service to start")

	// 初始化回调
	app.initCB.Exec(nil, app)

	// 解析启动参数
	pflag.Parse()

	// 合并启动参数配置
	startupConf := app.startupConf

	startupConf.AutomaticEnv()
	startupConf.BindPFlags(pflag.CommandLine)
	startupConf.SetConfigType(startupConf.GetString("conf.format"))
	startupConf.SetConfigFile(startupConf.GetString("conf.local_path"))

	if startupConf.ConfigFileUsed() != "" {
		if err := startupConf.ReadInConfig(); err != nil {
			panic(fmt.Errorf("read config file failed, %s", err))
		}
	}

	// 启动回调
	app.startingCB.Exec(nil, app)

	// 监听退出信号
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigChan
		cancel()
	}()

	// 启动服务
	wg := &sync.WaitGroup{}

	serviceNum := startupConf.GetStringMapString("startup.services")

	for name, pt := range app.servicePTs {
		pt.generic.setup(startupConf, name, pt.generic)
		pt.num, _ = strconv.Atoi(serviceNum[name])
	}

	for _, pt := range app.servicePTs {
		for i := 0; i < pt.num; i++ {
			wg.Add(1)
			go func(generic iServiceGeneric, no int) {
				defer wg.Done()
				<-generic.generate(ctx, no).Run()
			}(pt.generic, i)
		}
	}

	wg.Wait()

	// 结束回调
	app.terminatedCB.Exec(nil, app)
}

// GetStartupConf 获取启动参数配置
func (app *App) GetStartupConf() *viper.Viper {
	return app.startupConf
}
