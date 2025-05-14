/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package framework

import (
	"context"
	"errors"
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net"
	"net/http"
	_ "net/http/pprof"
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

type _ServicePT struct {
	generic iServiceGeneric
	num     int
}

// App 应用
type App struct {
	servicePTs               map[string]*_ServicePT
	startupConf              *viper.Viper
	initCB                   generic.Action1[*cobra.Command]
	startingCB, terminatedCB generic.Action1[*App]
}

func (app *App) lazyInit() {
	if app.servicePTs == nil {
		app.servicePTs = make(map[string]*_ServicePT)
	}
	if app.startupConf == nil {
		app.startupConf = viper.New()
	}
}

// Setup 安装服务泛化类型
func (app *App) Setup(name string, generic any) *App {
	app.lazyInit()

	if generic == nil {
		exception.Panicf("%w: %w: generic is nil", ErrFramework, core.ErrArgs)
	}

	svcGeneric, ok := generic.(iServiceGeneric)
	if !ok {
		svcGeneric = newServiceInstantiation(generic)
	}

	svcGeneric.init(app.startupConf, name, svcGeneric)

	app.servicePTs[name] = &_ServicePT{
		generic: svcGeneric,
		num:     1,
	}

	return app
}

// InitCB 初始化回调
func (app *App) InitCB(cb generic.Action1[*cobra.Command]) *App {
	app.initCB = cb
	return app
}

// StartingCB 启动回调
func (app *App) StartingCB(cb generic.Action1[*App]) *App {
	app.startingCB = cb
	return app
}

// TerminateCB 终止回调
func (app *App) TerminateCB(cb generic.Action1[*App]) *App {
	app.terminatedCB = cb
	return app
}

// Run 运行
func (app *App) Run() {
	app.lazyInit()

	cmd := &cobra.Command{
		Short: "Application for Launching Services",
		Run: func(cmd *cobra.Command, args []string) {
			// 加载启动参数
			app.initStartupConf(cmd)

			// 启动pprof
			app.initPProf()

			// 启动回调
			app.startingCB.UnsafeCall(app)

			// 主循环
			app.mainLoop()
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			// 结束回调
			app.terminatedCB.UnsafeCall(app)
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   true,
			DisableDescriptions: true,
		},
	}

	// 初始化参数
	app.initFlags(cmd)

	// 初始化回调
	app.initCB.UnsafeCall(cmd)

	// 开始运行
	if err := cmd.Execute(); err != nil {
		exception.Panicf("%w: %w", ErrFramework, err)
	}
}

// GetStartupConf 获取启动参数配置
func (app *App) GetStartupConf() *viper.Viper {
	return app.startupConf
}

func (app *App) initFlags(cmd *cobra.Command) {
	// 日志参数
	cmd.PersistentFlags().String("log.format", "console", "logging output format (json|console)")
	cmd.PersistentFlags().String("log.level", "info", "logging level")
	cmd.PersistentFlags().String("log.dir", "", "logging directory path")
	cmd.PersistentFlags().Int("log.size", 100*1024*1024, "log file splitting size")
	cmd.PersistentFlags().Bool("log.stdout", true, "logging output to stdout")
	cmd.PersistentFlags().Bool("log.development", false, "logging in development mode")
	cmd.PersistentFlags().Bool("log.service_info", false, "logging output service info")
	cmd.PersistentFlags().Bool("log.runtime_info", false, "logging output generic info")

	// 配置参数
	cmd.PersistentFlags().String("conf.env_prefix", "", "defines the prefix for environment variables")
	cmd.PersistentFlags().String("conf.local_path", "", "local config file path")
	cmd.PersistentFlags().String("conf.remote_provider", "", "remote config provider")
	cmd.PersistentFlags().String("conf.remote_endpoint", "", "remote config endpoint")
	cmd.PersistentFlags().String("conf.remote_path", "", "remote config file path")
	cmd.PersistentFlags().Bool("conf.auto_hotfix", false, "auto hotfix config")

	// nats参数
	cmd.PersistentFlags().String("nats.address", "localhost:4222", "nats address")
	cmd.PersistentFlags().String("nats.username", "", "nats auth username")
	cmd.PersistentFlags().String("nats.password", "", "nats auth password")

	// etcd参数
	cmd.PersistentFlags().String("etcd.address", "localhost:2379", "etcd address")
	cmd.PersistentFlags().String("etcd.username", "", "etcd auth username")
	cmd.PersistentFlags().String("etcd.password", "", "etcd auth password")

	// 分布式服务参数
	cmd.PersistentFlags().String("service.version", "v0.0.0", "service version info")
	cmd.PersistentFlags().StringToString("service.meta", map[string]string{}, "service meta info")
	cmd.PersistentFlags().Duration("service.ttl", 10*time.Second, "ttl for service keepalive")
	cmd.PersistentFlags().Duration("service.future_timeout", 3*time.Second, "timeout for future model of service interaction")
	cmd.PersistentFlags().Duration("service.dent_ttl", 10*time.Second, "ttl for distributed entity keepalive")
	cmd.PersistentFlags().Bool("service.auto_recover", false, "enable panic auto recover")

	// 启动的服务列表
	cmd.PersistentFlags().StringToString("startup.services", func() map[string]string {
		ret := map[string]string{}
		for name, pt := range app.servicePTs {
			ret[name] = strconv.Itoa(pt.num)
		}
		return ret
	}(), "instances required for each service to start")

	// pprof参数
	cmd.PersistentFlags().Bool("pprof.enable", false, "enable pprof")
	cmd.PersistentFlags().String("pprof.address", "0.0.0.0:6060", "pprof listening address")
}

func (app *App) initStartupConf(cmd *cobra.Command) {
	startupConf := app.startupConf

	// 合并启动参数
	startupConf.BindPFlags(cmd.Flags())

	// 合并环境变量
	startupConf.SetEnvPrefix(startupConf.GetString("conf.env_prefix"))
	startupConf.AutomaticEnv()

	// 加载本地配置文件
	localPath := startupConf.GetString("conf.local_path")

	if localPath != "" {
		startupConf.SetConfigFile(localPath)

		if err := startupConf.ReadInConfig(); err != nil {
			exception.Panicf("%w: read local config %q failed, %s", ErrFramework, localPath, err)
		}
	}

	// 加载远程配置文件
	remoteProvider := startupConf.GetString("conf.remote_provider")
	remoteEndpoint := startupConf.GetString("conf.remote_endpoint")
	remotePath := startupConf.GetString("conf.remote_path")

	if remoteProvider != "" {
		if err := startupConf.AddRemoteProvider(remoteProvider, remoteEndpoint, remotePath); err != nil {
			exception.Panicf(`%w: set remote config "%s - %s - %s" failed, %s`, ErrFramework, remoteProvider, remoteEndpoint, remotePath, err)
		}
		if err := startupConf.ReadRemoteConfig(); err != nil {
			exception.Panicf(`%w: read remote config "%s - %s - %s" failed, %s`, ErrFramework, remoteProvider, remoteEndpoint, remotePath, err)
		}
	}
}

func (app *App) initPProf() {
	if !app.GetStartupConf().GetBool("pprof.enable") {
		return
	}

	addr := app.GetStartupConf().GetString("pprof.address")

	_, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		exception.Panicf("%w: startup config [--pprof.address] = %q is invalid, %s", ErrFramework, addr, err)
	}

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
			exception.Panicf("%w: interrupt listening %q, %s", ErrFramework, addr, err)
		}
	}()
}

func (app *App) mainLoop() {
	// 监听退出信号
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigChan
		cancel()
	}()

	// 启动所有服务
	wg := &sync.WaitGroup{}

	serviceNum := app.startupConf.GetStringMapString("startup.services")

	for name, pt := range app.servicePTs {
		pt.num, _ = strconv.Atoi(serviceNum[name])
	}

	for _, pt := range app.servicePTs {
		for i := 0; i < pt.num; i++ {
			wg.Add(1)
			go func(svcGeneric iServiceGeneric, no int) {
				defer wg.Done()
				<-svcGeneric.generate(ctx, no).Run()
			}(pt.generic, i)
		}
	}

	// 等待运行结束
	wg.Wait()
}
