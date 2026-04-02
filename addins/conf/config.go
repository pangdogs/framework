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
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
)

// IConfig 配置接口
type IConfig interface {
	// AppConf 当前应用配置
	AppConf() *viper.Viper
	// ServiceConf 当前服务配置
	ServiceConf() *viper.Viper
}

// A 获取当前应用配置
func A(provider service.Context) *viper.Viper {
	return AddIn.Require(provider).AppConf()
}

// S 获取当前服务配置
func S(provider service.Context) *viper.Viper {
	return AddIn.Require(provider).ServiceConf()
}

func newConfig(settings ...option.Setting[ConfigOptions]) IConfig {
	return &_Config{
		options: option.New(With.Default(), settings...),
	}
}

type _Config struct {
	svcCtx      service.Context
	options     ConfigOptions
	appConf     *viper.Viper
	serviceConf *viper.Viper
}

// Init 初始化插件
func (c *_Config) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	c.svcCtx = svcCtx

	v := c.options.Vipper
	if v == nil {
		v = viper.New()
	}

	c.appConf = v
	c.serviceConf = v.Sub(svcCtx.Name())
}

// Shut 关闭插件
func (c *_Config) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))
}

// AppConf 当前应用程序配置
func (c *_Config) AppConf() *viper.Viper {
	return c.appConf
}

// ServiceConf 当前服务配置
func (c *_Config) ServiceConf() *viper.Viper {
	return c.serviceConf
}
