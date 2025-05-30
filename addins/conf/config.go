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
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"sync"
	"time"
)

// IConfig 配置接口
type IConfig interface {
	// AppConf 当前应用程序配置
	AppConf() *viper.Viper
	// ServiceConf 当前服务配置
	ServiceConf() *viper.Viper
	// Hotfix 热更新
	Hotfix() error
}

func newConfig(settings ...option.Setting[ConfigOptions]) IConfig {
	return &_Config{
		options: option.Make(With.Default(), settings...),
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
	log.Infof(svcCtx, "init addin %q", self.Name)

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
			log.Panicf(svcCtx, "read local config %q failed, %s", c.options.LocalPath, err)
		}

		log.Infof(svcCtx, "load local config %q config ok", c.options.LocalPath)
	}

	if remote {
		if err := c.startupConf.AddRemoteProvider(c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath); err != nil {
			log.Panicf(svcCtx, `set remote config "%s - %s - %s" failed, %s`, c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
		}
		if err := c.startupConf.ReadRemoteConfig(); err != nil {
			log.Panicf(svcCtx, `read remote config "%s - %s - %s" failed, %s`, c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
		}

		log.Infof(svcCtx, `load remote config "%s - %s - %s" ok`, c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath)
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

				log.Infof(svcCtx, "auto hotfix reload local config %q ok", c.options.LocalPath)
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
						log.Errorf(svcCtx, `auto hotfix watch remote config "%s - %s - %s" changes failed, %s`, c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
						continue
					}

					c.updateConf()

					log.Infof(svcCtx, `auto hotfix reload remote config "%s - %s - %s" ok`, c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath)
				}
			}()
		}
	}
}

// Shut 关闭插件
func (c *_Config) Shut(svcCtx service.Context) {
	log.Infof(svcCtx, "shut addin %q", self.Name)
}

// AppConf 当前应用程序配置
func (c *_Config) AppConf() *viper.Viper {
	return c.appConf
}

// ServiceConf 当前服务配置
func (c *_Config) ServiceConf() *viper.Viper {
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
			errs = append(errs, err)
		} else {
			update = true
		}
	}

	if remote {
		if err := c.startupConf.ReadRemoteConfig(); err != nil {
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

	serviceConf := appConf.Sub(c.svcCtx.GetName())
	if serviceConf == nil {
		serviceConf = viper.New()
	}

	c.appConf = appConf
	c.serviceConf = serviceConf
}
