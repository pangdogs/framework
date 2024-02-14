package app

import (
	"fmt"
	"git.golaxy.org/core"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func NewApp() *App {
	return &App{}
}

type _ServInfo struct {
	serv _IService
	num  int
}

type App struct {
	servInfos   map[string]*_ServInfo
	startupConf *viper.Viper
}

func (app *App) lazyInit() {
	if app.servInfos == nil {
		app.servInfos = make(map[string]*_ServInfo)
	}
	if app.startupConf == nil {
		app.startupConf = viper.New()
	}
}

func (app *App) Setup(name string, serv any, num ...int) *App {
	app.lazyInit()

	if serv == nil {
		panic(fmt.Errorf("%w: serv is nil", core.ErrArgs))
	}

	_serv, ok := serv.(_IService)
	if ok {
		panic(fmt.Errorf("%w: incorrect serv type", core.ErrArgs))
	}

	var _num int
	if len(num) > 0 {
		_num = num[0]
	}
	if _num < 0 {
		_num = 1
	}

	app.servInfos[name] = &_ServInfo{
		serv: _serv,
		num:  _num,
	}
	_serv.init(app, name, serv)

	return app
}

func (app *App) Run() {
	app.lazyInit()

	// 日志参数
	pflag.String("log.level", "info", "logging level")
	pflag.String("log.file", filepath.Join("./log/", strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))+".log"), "log file path")
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

	// 启动
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

	wg := &sync.WaitGroup{}

	for _, si := range app.servInfos {
		for i := 0; i < si.num; i++ {
			wg.Add(1)
			go func(serv _IService) {
				defer wg.Done()
				<-core.NewService(serv.generate()).Run()
			}(si.serv)
		}
	}

	wg.Wait()
}
