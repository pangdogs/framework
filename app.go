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
	_ "net/http/pprof"
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

// NewApp 创建应用
func NewApp() *App {
	app := &App{
		conf: viper.New(),
	}
	app.cmd = &cobra.Command{
		Short: "Application for Launching Services",
		Run: func(*cobra.Command, []string) {
			// 加载参数配置
			app.initConf()
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
			DisableDefaultCmd: true,
		},
	}
	return app
}

type _SupportedService struct {
	assembler iServiceAssembler
	count     int
}

// App 应用
type App struct {
	services                         generic.SliceMap[string, *_SupportedService]
	conf                             *viper.Viper
	cmd                              *cobra.Command
	initCB, startingCB, terminatedCB generic.Action1[*App]
}

// SetAssembler 设置服务实例装配器
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

// InitCB 初始化回调
func (app *App) InitCB(cb generic.Action1[*App]) *App {
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
	if app.services == nil {
		exception.Panicf("%w: services is nil", ErrFramework)
	}

	if app.conf == nil {
		exception.Panicf("%w: conf is nil", ErrFramework)
	}

	if app.cmd == nil {
		exception.Panicf("%w: cmd is nil", ErrFramework)
	}

	// 初始化启动参数
	app.initFlags()
	// 执行初始化回调
	app.initCB.UnsafeCall(app)

	// 开始运行
	if err := app.cmd.Execute(); err != nil {
		exception.Panicf("%w: %w", ErrFramework, err)
	}
}

// Conf 获取参数配置
func (app *App) Conf() *viper.Viper {
	return app.conf
}

// Cmd 获取启动命令
func (app *App) Cmd() *cobra.Command {
	return app.cmd
}

func (app *App) initFlags() {
	cmd := app.cmd

	// 日志参数
	cmd.PersistentFlags().String("log.level", zap.InfoLevel.String(), "log level: [debug|info|warn|error|dpanic|panic|fatal]")
	cmd.PersistentFlags().String("log.encoder", "development", "log encoder: [production|development]")
	cmd.PersistentFlags().String("log.format", "console", "log format: [console|json]")
	cmd.PersistentFlags().Bool("log.async", true, "enable async log writer")
	cmd.PersistentFlags().Int("log.buffer_size", 512*1024, "async log buffer size in bytes")
	cmd.PersistentFlags().Duration("log.flush_interval", time.Second, "async log flush interval, e.g. 1s")

	// 配置参数
	cmd.PersistentFlags().String("conf.env_prefix", "", "defines the prefix for environment variables")
	cmd.PersistentFlags().String("conf.local_path", "", "local config file path")
	cmd.PersistentFlags().String("conf.remote_provider", "", "remote config provider")
	cmd.PersistentFlags().String("conf.remote_endpoint", "", "remote config endpoint")
	cmd.PersistentFlags().String("conf.remote_path", "", "remote config file path")

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
		app.services.Each(func(name string, service *_SupportedService) {
			ret[name] = strconv.Itoa(service.count)
		})
		return ret
	}(), "instances required for each service to start")

	// pprof参数
	cmd.PersistentFlags().Bool("pprof.enable", false, "enable pprof")
	cmd.PersistentFlags().String("pprof.address", "0.0.0.0:6060", "pprof listening address")
}

func (app *App) initConf() {
	conf := app.conf

	// 合并启动参数
	conf.BindPFlags(app.cmd.Flags())

	// 合并环境变量
	conf.SetEnvPrefix(conf.GetString("conf.env_prefix"))
	conf.AutomaticEnv()

	// 加载本地配置文件
	localPath := conf.GetString("conf.local_path")

	if localPath != "" {
		conf.SetConfigFile(localPath)

		if err := conf.ReadInConfig(); err != nil {
			exception.Panicf("%w: read local config failed, path:%q, %s", ErrFramework, localPath, err)
		}
	}

	// 加载远程配置文件
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

	// 等待运行结束
	wg.Wait()
}
