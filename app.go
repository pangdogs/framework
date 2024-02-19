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
	"sync"
	"syscall"
	"time"
)

// NewApp 创建应用
func NewApp() *App {
	return &App{}
}

type _ServInfo struct {
	serv _IService
	num  int
}

// App 应用
type App struct {
	servInfos                     map[string]*_ServInfo
	startupConf                   *viper.Viper
	initCB, startingCB, stoppedCB generic.DelegateAction1[*App]
}

func (app *App) lazyInit() {
	if app.servInfos == nil {
		app.servInfos = make(map[string]*_ServInfo)
	}
	if app.startupConf == nil {
		app.startupConf = viper.New()
	}
}

// Setup 安装服务
func (app *App) Setup(name string, serv any, num ...int) *App {
	app.lazyInit()

	if serv == nil {
		panic(fmt.Errorf("%w: serv is nil", core.ErrArgs))
	}

	_serv, ok := serv.(_IService)
	if !ok {
		panic(fmt.Errorf("%w: incorrect serv type", core.ErrArgs))
	}

	var _num int
	if len(num) > 0 {
		_num = num[0]
	}
	if _num <= 0 {
		_num = 1
	}

	app.servInfos[name] = &_ServInfo{
		serv: _serv,
		num:  _num,
	}
	_serv.init(app.startupConf, name, serv)

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

// StoppedCB 停止回调
func (app *App) StoppedCB(cb generic.DelegateAction1[*App]) *App {
	app.stoppedCB = cb
	return app
}

// Run 回调
func (app *App) Run() {
	app.lazyInit()

	// 日志参数
	pflag.String("log.level", "info", "logging level")
	pflag.String("log.dir", "./log/", "logging directory path")
	pflag.Int("log.size", 100*1024*1024, "log file splitting size")
	pflag.Bool("log.stdout", false, "logging output to stdout")
	pflag.Bool("log.development", false, "logging in development mode")

	// 配置参数
	pflag.String("conf.format", "json", "config file format")
	pflag.String("conf.startup_path", "", "startup config file path")
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
	pflag.Duration("service.ttl", 10*time.Second, "ttl for service keepalive")
	pflag.Duration("service.future_timeout", 3*time.Second, "timeout for future model of service interaction")
	pflag.Duration("service.dent_ttl", 10*time.Second, "ttl for distributed entity keepalive")
	pflag.Bool("service.auto_recover", false, "enable panic auto recover")

	// 初始化回调
	app.initCB.Exec(nil, app)

	// 解析启动参数
	pflag.Parse()

	// 合并启动参数配置
	startupConf := app.startupConf

	startupConf.AutomaticEnv()
	startupConf.BindPFlags(pflag.CommandLine)
	startupConf.SetConfigType(startupConf.GetString("conf.format"))
	startupConf.SetConfigFile(startupConf.GetString("conf.startup_path"))

	if startupConf.ConfigFileUsed() != "" {
		if err := startupConf.ReadInConfig(); err != nil {
			panic(fmt.Errorf("read startup config file failed, %s", err))
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

	for _, si := range app.servInfos {
		for i := 0; i < si.num; i++ {
			wg.Add(1)
			go func(serv _IService) {
				defer wg.Done()
				<-serv.generate(ctx).Run()
			}(si.serv)
		}
	}

	wg.Wait()

	// 结束回调
	app.stoppedCB.Exec(nil, app)
}

// GetStartupConf 获取启动参数配置
func (app *App) GetStartupConf() *viper.Viper {
	return app.startupConf
}
