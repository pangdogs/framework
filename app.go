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
	"sync/atomic"
	"syscall"
	"time"
)

// NewApp 创建应用
func NewApp() *App {
	return &App{
		startupConf: viper.New(),
		servicePTs:  map[string]*_ServicePT{},
	}
}

type _ServicePT struct {
	generic iServiceGeneric
	num     int
}

// App 应用
type App struct {
	isRunning                        atomic.Bool
	servicePTs                       map[string]*_ServicePT
	startupConf                      *viper.Viper
	startupCmd                       *cobra.Command
	initCB, startingCB, terminatedCB generic.Action1[*App]
}

// Setup 安装服务泛化类型
func (app *App) Setup(name string, generic any) *App {
	if app.isRunning.Load() {
		exception.Panicf("%w: already running", ErrFramework)
	}

	if generic == nil {
		exception.Panicf("%w: %w: generic is nil", ErrFramework, core.ErrArgs)
	}

	svcGeneric, ok := generic.(iServiceGeneric)
	if !ok {
		svcGeneric = newServiceInstantiation(generic)
	}

	app.servicePTs[name] = &_ServicePT{
		generic: svcGeneric,
		num:     1,
	}

	return app
}

// InitCB 初始化回调
func (app *App) InitCB(cb generic.Action1[*App]) *App {
	if app.isRunning.Load() {
		exception.Panicf("%w: already running", ErrFramework)
	}
	app.initCB = cb
	return app
}

// StartingCB 启动回调
func (app *App) StartingCB(cb generic.Action1[*App]) *App {
	if app.isRunning.Load() {
		exception.Panicf("%w: already running", ErrFramework)
	}
	app.startingCB = cb
	return app
}

// TerminateCB 终止回调
func (app *App) TerminateCB(cb generic.Action1[*App]) *App {
	if app.isRunning.Load() {
		exception.Panicf("%w: already running", ErrFramework)
	}
	app.terminatedCB = cb
	return app
}

// Run 运行
func (app *App) Run() {
	if !app.isRunning.CompareAndSwap(false, true) {
		exception.Panicf("%w: already running", ErrFramework)
	}

	// 初始化启动命令
	app.startupCmd = &cobra.Command{
		Short: "Application for Launching Services",
		Run: func(*cobra.Command, []string) {
			// 加载启动参数配置
			app.initStartupConf()
			// 加载pprof
			app.initPProf()
			// 执行启动回调
			app.startingCB.UnsafeCall(app)
			// 主循环
			app.mainLoop()
			// 执行结束回调
			app.terminatedCB.UnsafeCall(app)
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   true,
			DisableDescriptions: true,
		},
	}

	// 初始化已安装的服务泛化类型
	for name, servicePT := range app.servicePTs {
		servicePT.generic.init(app.startupConf, app.startupCmd, name, servicePT.generic)
	}

	// 初始化启动参数
	app.initFlags()
	// 执行初始化回调
	app.initCB.UnsafeCall(app)

	// 开始运行
	if err := app.startupCmd.Execute(); err != nil {
		exception.Panicf("%w: %w", ErrFramework, err)
	}
}

// GetStartupConf 获取启动参数配置
func (app *App) GetStartupConf() *viper.Viper {
	return app.startupConf
}

// GetStartupCmd 获取启动命令
func (app *App) GetStartupCmd() *cobra.Command {
	return app.startupCmd
}

func (app *App) initFlags() {
	startupCmd := app.startupCmd

	// 日志参数
	startupCmd.PersistentFlags().String("log.format", "console", "logging output format (json|console)")
	startupCmd.PersistentFlags().String("log.level", "info", "logging level")
	startupCmd.PersistentFlags().String("log.dir", "", "logging directory path")
	startupCmd.PersistentFlags().Int("log.size", 100*1024*1024, "log file splitting size")
	startupCmd.PersistentFlags().Bool("log.stdout", true, "logging output to stdout")
	startupCmd.PersistentFlags().Bool("log.development", false, "logging in development mode")
	startupCmd.PersistentFlags().Bool("log.service_info", false, "logging output service info")
	startupCmd.PersistentFlags().Bool("log.runtime_info", false, "logging output generic info")

	// 配置参数
	startupCmd.PersistentFlags().String("conf.env_prefix", "", "defines the prefix for environment variables")
	startupCmd.PersistentFlags().String("conf.local_path", "", "local config file path")
	startupCmd.PersistentFlags().String("conf.remote_provider", "", "remote config provider")
	startupCmd.PersistentFlags().String("conf.remote_endpoint", "", "remote config endpoint")
	startupCmd.PersistentFlags().String("conf.remote_path", "", "remote config file path")
	startupCmd.PersistentFlags().Bool("conf.auto_hotfix", false, "auto hotfix config")

	// nats参数
	startupCmd.PersistentFlags().String("nats.address", "localhost:4222", "nats address")
	startupCmd.PersistentFlags().String("nats.username", "", "nats auth username")
	startupCmd.PersistentFlags().String("nats.password", "", "nats auth password")

	// etcd参数
	startupCmd.PersistentFlags().String("etcd.address", "localhost:2379", "etcd address")
	startupCmd.PersistentFlags().String("etcd.username", "", "etcd auth username")
	startupCmd.PersistentFlags().String("etcd.password", "", "etcd auth password")

	// 分布式服务参数
	startupCmd.PersistentFlags().String("service.version", "v0.0.0", "service version info")
	startupCmd.PersistentFlags().StringToString("service.meta", map[string]string{}, "service meta info")
	startupCmd.PersistentFlags().Duration("service.ttl", 10*time.Second, "ttl for service keepalive")
	startupCmd.PersistentFlags().Duration("service.future_timeout", 3*time.Second, "timeout for future model of service interaction")
	startupCmd.PersistentFlags().Duration("service.dent_ttl", 10*time.Second, "ttl for distributed entity keepalive")
	startupCmd.PersistentFlags().Bool("service.auto_recover", false, "enable panic auto recover")

	// 启动的服务列表
	startupCmd.PersistentFlags().StringToString("startup.services", func() map[string]string {
		ret := map[string]string{}
		for name, pt := range app.servicePTs {
			ret[name] = strconv.Itoa(pt.num)
		}
		return ret
	}(), "instances required for each service to start")

	// pprof参数
	startupCmd.PersistentFlags().Bool("pprof.enable", false, "enable pprof")
	startupCmd.PersistentFlags().String("pprof.address", "0.0.0.0:6060", "pprof listening address")
}

func (app *App) initStartupConf() {
	startupConf := app.startupConf

	// 合并启动参数
	startupConf.BindPFlags(app.startupCmd.Flags())

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
		exception.Panicf("%w: invalid pprof address %q, %s", ErrFramework, addr, err)
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

	for name, servicePT := range app.servicePTs {
		servicePT.num, _ = strconv.Atoi(serviceNum[name])
	}

	for _, servicePT := range app.servicePTs {
		for i := 0; i < servicePT.num; i++ {
			wg.Add(1)
			go func(svcGeneric iServiceGeneric, no int) {
				defer wg.Done()
				<-svcGeneric.generate(ctx, no).Run()
			}(servicePT.generic, i)
		}
	}

	// 等待运行结束
	wg.Wait()
}
