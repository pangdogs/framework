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

package conf

import (
	"errors"
	"sync"
	"time"

	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
)

// IConfig 配置接口
type IConfig interface {
	// App 当前应用程序配置
	App() *viper.Viper
	// Service 当前服务配置
	Service() *viper.Viper
	// Hotfix 热更新
	Hotfix() error
}

// A 获取当前应用程序配置
func A(provider service.Context) *viper.Viper {
	return AddIn.Require(provider).App()
}

// S 获取当前服务配置
func S(provider service.Context) *viper.Viper {
	return AddIn.Require(provider).Service()
}

func newConfig(settings ...option.Setting[ConfigOptions]) IConfig {
	return &_Config{
		options: option.New(With.Default(), settings...),
	}
}

type _Config struct {
	svcCtx                            service.Context
	options                           ConfigOptions
	startupConf, appConf, serviceConf *viper.Viper
	updateMutex                       sync.Mutex
}

// Init 初始化插件
func (c *_Config) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	c.svcCtx = svcCtx
	c.startupConf = viper.New()

	for k, v := range c.options.Defaults {
		c.startupConf.SetDefault(k, v)
	}

	if c.options.Flags != nil {
		c.startupConf.BindPFlags(c.options.Flags)
	}

	if c.options.AutomaticEnv {
		c.startupConf.SetEnvPrefix(c.options.EnvPrefix)
		c.startupConf.AutomaticEnv()
	}

	local := c.options.LocalPath != ""
	remote := c.options.RemoteProvider != ""

	if local {
		c.startupConf.SetConfigFile(c.options.LocalPath)

		if err := c.startupConf.ReadInConfig(); err != nil {
			log.L(svcCtx).Panic("read local config failed", zap.String("path", c.options.LocalPath), zap.Error(err))
		}

		log.L(svcCtx).Info("load local config ok", zap.String("path", c.options.LocalPath))
	}

	if remote {
		if err := c.startupConf.AddRemoteProvider(c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath); err != nil {
			log.L(svcCtx).Panic("add remote provider failed",
				zap.String("provider", c.options.RemoteProvider),
				zap.String("endpoint", c.options.RemoteEndpoint),
				zap.String("path", c.options.RemotePath),
				zap.Error(err))
		}
		if err := c.startupConf.ReadRemoteConfig(); err != nil {
			log.L(svcCtx).Panic("read remote config failed",
				zap.String("provider", c.options.RemoteProvider),
				zap.String("endpoint", c.options.RemoteEndpoint),
				zap.String("path", c.options.RemotePath),
				zap.Error(err))
		}

		log.L(svcCtx).Info("load remote config ok",
			zap.String("provider", c.options.RemoteProvider),
			zap.String("endpoint", c.options.RemoteEndpoint),
			zap.String("path", c.options.RemotePath))
	}

	c.updateConf()

	if c.options.AutoHotFix {
		if local {
			c.startupConf.OnConfigChange(func(in fsnotify.Event) {
				select {
				case <-c.svcCtx.Done():
					return
				default:
				}

				c.updateConf()

				log.L(c.svcCtx).Info("auto hotfix reload local config ok", zap.String("path", c.options.LocalPath))
			})
			c.startupConf.WatchConfig()
		}
		if remote {
			go func() {
				for {
					time.Sleep(c.options.AutoHotFixRemoteCheckingIntervalTime)

					select {
					case <-c.svcCtx.Done():
						return
					default:
					}

					if err := c.startupConf.WatchRemoteConfig(); err != nil {
						log.L(c.svcCtx).Error("auto hotfix watch remote config failed",
							zap.String("provider", c.options.RemoteProvider),
							zap.String("endpoint", c.options.RemoteEndpoint),
							zap.String("path", c.options.RemotePath),
							zap.Error(err))
						continue
					}

					c.updateConf()

					log.L(c.svcCtx).Info("auto hotfix reload remote config ok",
						zap.String("provider", c.options.RemoteProvider),
						zap.String("endpoint", c.options.RemoteEndpoint),
						zap.String("path", c.options.RemotePath))
				}
			}()
		}
	}
}

// Shut 关闭插件
func (c *_Config) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))
}

// App 当前应用程序配置
func (c *_Config) App() *viper.Viper {
	return c.appConf
}

// Service 当前服务配置
func (c *_Config) Service() *viper.Viper {
	return c.serviceConf
}

// Hotfix 热更新
func (c *_Config) Hotfix() error {
	local := c.options.LocalPath != ""
	remote := c.options.RemoteProvider != ""

	var errs []error
	var update bool

	if local {
		if err := c.startupConf.ReadInConfig(); err != nil {
			log.L(c.svcCtx).Error("read local config failed", zap.String("path", c.options.LocalPath), zap.Error(err))
			errs = append(errs, err)
		} else {
			update = true
		}
	}

	if remote {
		if err := c.startupConf.ReadRemoteConfig(); err != nil {
			log.L(c.svcCtx).Error("read remote config failed",
				zap.String("provider", c.options.RemoteProvider),
				zap.String("endpoint", c.options.RemoteEndpoint),
				zap.String("path", c.options.RemotePath),
				zap.Error(err))
			errs = append(errs, err)
		} else {
			update = true
		}
	}

	if update {
		c.updateConf()
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (c *_Config) updateConf() {
	c.updateMutex.Lock()
	defer c.updateMutex.Unlock()

	appConf := viper.New()
	appConf.MergeConfigMap(c.startupConf.AllSettings())

	serviceConf := appConf.Sub(c.svcCtx.Name())
	if serviceConf == nil {
		serviceConf = viper.New()
	}

	c.appConf = appConf
	c.serviceConf = serviceConf
}
