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
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote" // 注册 Viper 远程配置后端。
	"go.uber.org/zap"
)

// IConfig 提供应用级配置及当前服务的配置子树。
type IConfig interface {
	// AppConf 返回完整应用配置。
	AppConf() *viper.Viper
	// ServiceConf 返回以服务名为键的配置子树；不存在时可能返回 nil。
	ServiceConf() *viper.Viper
}

// A 返回 provider 所属服务的完整应用配置；配置 add-in 未安装时会 panic。
func A(provider service.Context) *viper.Viper {
	return AddIn.Require(provider).AppConf()
}

// S 返回 provider 所属服务的配置子树；配置 add-in 未安装时会 panic。
func S(provider service.Context) *viper.Viper {
	return AddIn.Require(provider).ServiceConf()
}

func newConfig(settings ...option.Setting[ConfigOptions]) IConfig {
	return &_Config{
		options: option.New(With.Default(), settings...),
	}
}

type _Config struct {
	options     ConfigOptions
	appConf     *viper.Viper
	serviceConf *viper.Viper
}

// Init 采用配置的 Viper 实例（未提供时新建空实例），并提取当前服务名对应的配置子树。
func (c *_Config) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	v := c.options.Vipper
	if v == nil {
		v = viper.New()
	}

	c.appConf = v
	c.serviceConf = v.Sub(svcCtx.Name())
}

// RetainAfterTermination 使配置 add-in 在 Service 终止后继续可用。
func (*_Config) RetainAfterTermination() {}

// AppConf 返回 Init 绑定的完整应用配置；调用方的修改会直接作用于共享实例。
func (c *_Config) AppConf() *viper.Viper {
	return c.appConf
}

// ServiceConf 返回当前服务的共享配置子树；对应配置不存在时返回 nil。
func (c *_Config) ServiceConf() *viper.Viper {
	return c.serviceConf
}
