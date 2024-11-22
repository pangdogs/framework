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
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"time"
)

// IConfig 配置接口
type IConfig interface {
	IVisitConf
	Whole() IVisitConf
}

func newConfig(settings ...option.Setting[ConfigOptions]) IConfig {
	return &_Config{
		options: option.Make(With.Default(), settings...),
	}
}

type _Config struct {
	options ConfigOptions
	*_VisitConf
	whole *_VisitConf
}

// Init 初始化插件
func (c *_Config) Init(svcCtx service.Context, _ runtime.Context) {
	log.Infof(svcCtx, "init plugin %q", self.Name)

	vp := viper.New()
	vp.SetConfigType(c.options.Format)

	for k, v := range c.options.DefaultKVs {
		vp.SetDefault(k, v)
	}

	if c.options.MergeEnv {
		vp.AutomaticEnv()
	}

	if c.options.MergeConf != nil {
		vp.MergeConfigMap(c.options.MergeConf.AllSettings())
	}

	local := c.options.LocalPath != ""
	remote := c.options.RemoteProvider != ""

	if local {
		vp.SetConfigFile(c.options.LocalPath)

		if err := vp.MergeInConfig(); err != nil {
			log.Panicf(svcCtx, "read local config %q failed, %s", c.options.LocalPath, err)
		}

		log.Infof(svcCtx, "load local config %q config ok", c.options.LocalPath)
	}

	if remote {
		if err := vp.AddRemoteProvider(c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath); err != nil {
			log.Panicf(svcCtx, "set remote config [%q, %q, %q] failed, %s", c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
		}
		if err := vp.ReadRemoteConfig(); err != nil {
			log.Panicf(svcCtx, "read remote config [%q, %q, %q] failed, %s", c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
		}

		log.Infof(svcCtx, "load remote config [%q, %q, %q] ok", c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath)
	}

	subVp := vp.Sub(svcCtx.GetName())
	if subVp == nil {
		subVp = viper.New()
	}
	c._VisitConf = &_VisitConf{
		Viper: subVp,
	}
	c.whole = &_VisitConf{
		Viper: vp,
	}

	if c.options.AutoHotFix {
		if local {
			vp.OnConfigChange(func(in fsnotify.Event) {
				subVp := vp.Sub(svcCtx.GetName())
				if subVp == nil {
					subVp = viper.New()
				}
				c._VisitConf.Viper = subVp

				log.Infof(svcCtx, "reload local config %q ok", c.options.LocalPath)
			})
			vp.WatchConfig()
		}
		if remote {
			go func() {
				for {
					time.Sleep(time.Second * 3)

					err := vp.WatchRemoteConfig()
					if err != nil {
						log.Errorf(svcCtx, "watch remote config [%q, %q, %q] changes failed, %s", c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath, err)
						continue
					}

					subVp := vp.Sub(svcCtx.GetName())
					if subVp == nil {
						subVp = viper.New()
					}
					c._VisitConf.Viper = subVp

					log.Infof(svcCtx, "reload remote config [%q, %q, %q] ok", c.options.RemoteProvider, c.options.RemoteEndpoint, c.options.RemotePath)
				}
			}()
		}
	}
}

// Shut 关闭插件
func (c *_Config) Shut(svcCtx service.Context, _ runtime.Context) {
	log.Infof(svcCtx, "shut plugin %q", self.Name)
}

func (c *_Config) Whole() IVisitConf {
	return c.whole
}
