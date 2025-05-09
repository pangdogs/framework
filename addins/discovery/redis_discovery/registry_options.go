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

package redis_discovery

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
	"github.com/redis/go-redis/v9"
	"net"
	"strings"
	"time"
)

// RegistryOptions 所有选项
type RegistryOptions struct {
	RedisClient    *redis.Client
	RedisConfig    *redis.Options
	RedisURL       string
	KeyPrefix      string
	WatchChanSize  int
	TTL            time.Duration
	CustomUsername string
	CustomPassword string
	CustomAddress  string
	CustomDB       int
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		With.RedisClient(nil).Apply(options)
		With.RedisConfig(nil).Apply(options)
		With.RedisURL("").Apply(options)
		With.KeyPrefix("golaxy:services:").Apply(options)
		With.WatchChanSize(128).Apply(options)
		With.TTL(10 * time.Second).Apply(options)
		With.CustomAuth("", "").Apply(options)
		With.CustomAddress("127.0.0.1:6379").Apply(options)
		With.CustomDB(0).Apply(options)
	}
}

// RedisClient redis客户端，1st优先使用
func (_Option) RedisClient(cli *redis.Client) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.RedisClient = cli
	}
}

// RedisConfig redis配置，2nd优先使用
func (_Option) RedisConfig(conf *redis.Options) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.RedisConfig = conf
	}
}

// RedisURL redis连接url，3rd优先使用
func (_Option) RedisURL(url string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.RedisURL = url
	}
}

// KeyPrefix 所有key的前缀
func (_Option) KeyPrefix(prefix string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		options.KeyPrefix = prefix
	}
}

// WatchChanSize 监控服务变化的channel大小
func (_Option) WatchChanSize(size int) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		if size < 0 {
			exception.Panicf("registry: %w: option WatchChanSize can't be set to a value less than 0", core.ErrArgs)
		}
		options.WatchChanSize = size
	}
}

// TTL 默认TTL
func (_Option) TTL(ttl time.Duration) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		if ttl < 3*time.Second {
			exception.Panicf("registry: %w: option TTL can't be set to a value less than 3 second", core.ErrArgs)
		}
		options.TTL = ttl
	}
}

// CustomAuth 自定义设置redis鉴权信息
func (_Option) CustomAuth(username, password string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddress 自定义设置redis服务地址
func (_Option) CustomAddress(addr string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			exception.Panicf("registry: %w: %w", core.ErrArgs, err)
		}
		options.CustomAddress = addr
	}
}

// CustomDB 自定义设置redis db
func (_Option) CustomDB(db int) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.CustomDB = db
	}
}
