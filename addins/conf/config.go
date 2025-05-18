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
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"time"
)

// IConfig 配置接口
type IConfig interface {
	// AppConf 当前应用程序配置
	AppConf() *viper.Viper
	// ServiceConf 当前服务配置
	ServiceConf() *viper.Viper
}

func newConfig(settings ...option.Setting[ConfigOptions]) IConfig {
	return &_Config{
		options: option.Make(With.Default(), settings...),
	}
}

type _Config struct {
	options              ConfigOptions
	appConf, serviceConf *viper.Viper
}

// Init 初始化插件
func (c *_Config) Init(svcCtx service.Context) {
	log.Infof(svcCtx, "init addin %q", self.Name)

	startupConf := viper.New()

	for k, v := range c.options.Defaults {
		startupConf.SetDefault(k, v)
	}

	if c.options.Flags != nil {
		startupConf.BindPFlags(c.options.Flags)
	}

	if c.options.AutomaticEnv {
		startupConf.SetEnvPrefix(c.options.EnvPrefix)
		startupConf.AutomaticEnv()
	}

	local := c.options.LocalPath != ""
	remote := c.options.RemoteProvider != ""

	if local {
		startupConf.SetConfigFile(c.options.LocalPath)

		if err := startupConf.MergeInConfig(); err != nil {
			log.Panicf(svcCtx, "read local config %q failed, %s", c.options.LocalPath, err)
		}

		log.Infof(svcCtx, "load local config %q config ok", c.options.LocalPath)
	}

	if remote {
		if err := startupConf.AddRemoteProvider(c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath); err != nil {
			log.Panicf(svcCtx, `set remote config "%s - %s - %s" failed, %s`, c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
		}
		if err := startupConf.ReadRemoteConfig(); err != nil {
			log.Panicf(svcCtx, `read remote config "%s - %s - %s" failed, %s`, c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
		}

		log.Infof(svcCtx, `load remote config "%s - %s - %s" ok`, c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath)
	}

	c.updateConf(startupConf, svcCtx.GetName())

	if c.options.AutoHotFix {
		if local {
			startupConf.OnConfigChange(func(in fsnotify.Event) {
				c.updateConf(startupConf, svcCtx.GetName())
				log.Infof(svcCtx, "reload local config %q ok", c.options.LocalPath)
			})
			startupConf.WatchConfig()
		}
		if remote {
			go func() {
				for {
					time.Sleep(time.Second * 3)

					if err := startupConf.WatchRemoteConfig(); err != nil {
						log.Errorf(svcCtx, `watch remote config "%s - %s - %s" changes failed, %s`, c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
						continue
					}

					c.updateConf(startupConf, svcCtx.GetName())

					log.Infof(svcCtx, `reload remote config "%s - %s - %s" ok`, c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath)
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

func (c *_Config) updateConf(startupConf *viper.Viper, svcName string) {
	appConf := viper.New()
	appConf.MergeConfigMap(startupConf.AllSettings())

	serviceConf := appConf.Sub(svcName)
	if serviceConf == nil {
		serviceConf = viper.New()
	}

	c.appConf = appConf
	c.serviceConf = serviceConf
}
