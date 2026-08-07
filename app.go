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
	"net"
	"net/http"
	_ "net/http/pprof" // 注册 pprof 的默认 HTTP 处理器，供 initPProf 启动的服务使用。
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// NewApp 创建一个尚未运行的应用，并初始化命令行、配置及服务装配容器。
func NewApp() *App {
	app := &App{
		conf: viper.New(),
	}
	app.cmd = &cobra.Command{
		Short: "Application for Launching Services",
		Run: func(*cobra.Command, []string) {
			// Cobra 入口依次合并配置、启动辅助服务、运行服务副本并执行生命周期回调。
			app.initConf()
			app.initPProf()
			app.startingCB.UnsafeCall(app)
			app.mainLoop()
			app.terminatedCB.UnsafeCall(app)
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	return app
}

type _SupportedService struct {
	assembler iServiceAssembler
	count     int
}

// App 负责装载应用配置、启动已注册的服务副本，并等待所有服务停止。
// App 的配置和装配方法应在 Run 前由同一 goroutine 调用。
type App struct {
	services                         generic.SliceMap[string, *_SupportedService]
	conf                             *viper.Viper
	cmd                              *cobra.Command
	initCB, startingCB, terminatedCB generic.Action1[*App]
	initOnce                         bool
}

// SetAssembler 注册名为 name 的服务装配器。
// assembler 可以是自定义装配器，也可以是实现 IService 的实例或 reflect.Type；
// 后两者会被转换为按需创建实例的默认装配器。同名注册会替换此前的装配器。
func (app *App) SetAssembler(name string, assembler any) *App {
	if app.conf == nil {
		exception.Panicf("%w: conf is nil", ErrFramework)
	}

	if app.cmd == nil {
		exception.Panicf("%w: cmd is nil", ErrFramework)
	}

	if assembler == nil {
		exception.Panicf("%w: %w: assembler is nil", ErrFramework, core.ErrArgs)
	}

	assemblerInst, ok := assembler.(iServiceAssembler)
	if !ok {
		assemblerInst = newServiceInstantiator(assembler)
	}
	assemblerInst.init(app.conf, app.cmd, name, assemblerInst)

	app.services.Add(name,
		&_SupportedService{
			assembler: assemblerInst,
			count:     1,
		},
	)

	return app
}

// InitCB 设置命令初始化回调。
// 回调在内置参数注册后、首次执行命令前调用一次，可用于补充 Cobra 参数或子命令。
func (app *App) InitCB(cb generic.Action1[*App]) *App {
	app.initCB = cb
	return app
}

// StartingCB 设置应用启动回调。
// 回调在配置与可选 pprof 服务初始化完成后、启动服务副本前执行。
func (app *App) StartingCB(cb generic.Action1[*App]) *App {
	app.startingCB = cb
	return app
}

// TerminateCB 设置应用终止回调。
// 回调在全部服务副本停止后执行。
func (app *App) TerminateCB(cb generic.Action1[*App]) *App {
	app.terminatedCB = cb
	return app
}

// Run 初始化并执行应用命令，随后启动配置指定的服务副本。
// Cobra 返回执行错误时 Run 会 panic；同一个 App 不应并发调用 Run。
func (app *App) Run() {
	if app.conf == nil {
		exception.Panicf("%w: conf is nil", ErrFramework)
	}

	if app.cmd == nil {
		exception.Panicf("%w: cmd is nil", ErrFramework)
	}

	if !app.initOnce {
		app.initOnce = true
		// Cmd 与 Run 共用一次初始化，确保内置参数和 InitCB 不会重复注册。
		app.initFlags()
		app.initCB.UnsafeCall(app)
	}

	// Execute 最终进入 NewApp 注册的 Run 回调。
	if err := app.cmd.Execute(); err != nil {
		exception.Panicf("%w: %w", ErrFramework, err)
	}
}

// Conf 返回应用持有的 Viper 配置实例。
func (app *App) Conf() *viper.Viper {
	return app.conf
}

// Cmd 返回应用的根 Cobra 命令。
// 首次调用会注册内置参数并执行 InitCB；该初始化与 Run 共享一次性状态。
func (app *App) Cmd() *cobra.Command {
	if app.conf == nil {
		exception.Panicf("%w: conf is nil", ErrFramework)
	}

	if app.cmd == nil {
		exception.Panicf("%w: cmd is nil", ErrFramework)
	}

	if !app.initOnce {
		app.initOnce = true
		// 允许调用方先通过 Cmd 补充参数，再在之后调用 Run。
		app.initFlags()
		app.initCB.UnsafeCall(app)
	}

	return app.cmd
}

func (app *App) initFlags() {
	cmd := app.cmd

	// 日志参数。
	cmd.PersistentFlags().String("log.level", zap.InfoLevel.String(), "log level: [debug|info|warn|error|dpanic|panic|fatal]")
	cmd.PersistentFlags().String("log.encoder", "development", "log encoder: [production|development]")
	cmd.PersistentFlags().String("log.format", "console", "log format: [console|json]")
	cmd.PersistentFlags().Bool("log.async", true, "enable async log writer")
	cmd.PersistentFlags().Int("log.buffer_size", 512*1024, "async log buffer size in bytes")
	cmd.PersistentFlags().Duration("log.flush_interval", time.Second, "async log flush interval, e.g. 1s")

	// 配置参数。
	cmd.PersistentFlags().String("conf.env_prefix", "", "defines the prefix for environment variables")
	cmd.PersistentFlags().String("conf.local_path", "", "local config file path")
	cmd.PersistentFlags().String("conf.remote_provider", "", "remote config provider")
	cmd.PersistentFlags().String("conf.remote_endpoint", "", "remote config endpoint")
	cmd.PersistentFlags().String("conf.remote_path", "", "remote config file path")

	// NATS 参数。
	cmd.PersistentFlags().String("nats.address", "localhost:4222", "nats address")
	cmd.PersistentFlags().String("nats.username", "", "nats auth username")
	cmd.PersistentFlags().String("nats.password", "", "nats auth password")

	// ETCD 参数。
	cmd.PersistentFlags().String("etcd.address", "localhost:2379", "etcd address")
	cmd.PersistentFlags().String("etcd.username", "", "etcd auth username")
	cmd.PersistentFlags().String("etcd.password", "", "etcd auth password")

	// 分布式服务参数。
	cmd.PersistentFlags().String("service.version", "v0.0.0", "service version info")
	cmd.PersistentFlags().StringToString("service.meta", map[string]string{}, "service meta info")
	cmd.PersistentFlags().Duration("service.ttl", 10*time.Second, "ttl for service keepalive")
	cmd.PersistentFlags().Duration("service.future_timeout", 3*time.Second, "timeout for future model of service interaction")
	cmd.PersistentFlags().Duration("service.dent_ttl", 10*time.Second, "ttl for distributed entity keepalive")
	cmd.PersistentFlags().Bool("service.auto_recover", false, "enable panic auto recover")

	// 各类服务默认启动的副本数。
	cmd.PersistentFlags().StringToString("startup.services", func() map[string]string {
		ret := map[string]string{}
		app.services.Each(func(name string, service *_SupportedService) {
			ret[name] = strconv.Itoa(service.count)
		})
		return ret
	}(), "instances required for each service to start")

	// pprof 参数。
	cmd.PersistentFlags().Bool("pprof.enable", false, "enable pprof")
	cmd.PersistentFlags().String("pprof.address", "0.0.0.0:6060", "pprof listening address")
}

func (app *App) initConf() {
	conf := app.conf

	// 命令行参数参与 Viper 的统一优先级解析。
	conf.BindPFlags(app.cmd.Flags())

	// 环境变量在设置可选前缀后自动参与查找。
	conf.SetEnvPrefix(conf.GetString("conf.env_prefix"))
	conf.AutomaticEnv()

	// 本地配置文件为空时跳过读取。
	localPath := conf.GetString("conf.local_path")

	if localPath != "" {
		conf.SetConfigFile(localPath)

		if err := conf.ReadInConfig(); err != nil {
			exception.Panicf("%w: read local config failed, path:%q, %s", ErrFramework, localPath, err)
		}
	}

	// 配置了 provider 时，从指定端点和路径读取远程配置。
	remoteProvider := conf.GetString("conf.remote_provider")
	remoteEndpoint := conf.GetString("conf.remote_endpoint")
	remotePath := conf.GetString("conf.remote_path")

	if remoteProvider != "" {
		if err := conf.AddRemoteProvider(remoteProvider, remoteEndpoint, remotePath); err != nil {
			exception.Panicf(`%w: set remote config failed, provider:%q, endpoint:%q, path:%q, %s`, ErrFramework, remoteProvider, remoteEndpoint, remotePath, err)
		}
		if err := conf.ReadRemoteConfig(); err != nil {
			exception.Panicf(`%w: read remote config failed, provider:%q, endpoint:%q, path:%q, %s`, ErrFramework, remoteProvider, remoteEndpoint, remotePath, err)
		}
	}
}

func (app *App) initPProf() {
	if !app.Conf().GetBool("pprof.enable") {
		return
	}

	addr := app.Conf().GetString("pprof.address")

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
	// 首个退出信号取消共享上下文，通知全部服务副本停止。
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigChan
		cancel()
	}()

	// 配置值覆盖注册时的默认副本数；无效数量按零处理。
	wg := &sync.WaitGroup{}

	bootstrap := app.conf.GetStringMapString("startup.services")

	app.services.Each(func(name string, service *_SupportedService) {
		service.count, _ = strconv.Atoi(bootstrap[name])
	})

	app.services.Each(func(name string, service *_SupportedService) {
		for i := 0; i < service.count; i++ {
			wg.Add(1)
			go func(assembler iServiceAssembler, replicaNo int) {
				defer wg.Done()
				<-assembler.assemble(ctx, replicaNo).Run().Done()
			}(service.assembler, i)
		}
	})

	// 所有服务副本的 Done 完成后才返回 Cobra 入口。
	wg.Wait()
}
