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

package dsync_redis

import (
	"net"
	"strings"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
	"github.com/redis/go-redis/v9"
)

// RedisSyncOptions 所有选项
type RedisSyncOptions struct {
	RedisClient    *redis.Client
	RedisConfig    *redis.Options
	RedisURL       string
	KeyPrefix      string
	CustomUsername string
	CustomPassword string
	CustomAddress  string
	CustomDB       int
}

var With _RedisSyncOption

type _RedisSyncOption struct{}

// Default 默认选项
func (_RedisSyncOption) Default() option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		With.RedisClient(nil).Apply(options)
		With.RedisConfig(nil).Apply(options)
		With.RedisURL("").Apply(options)
		With.KeyPrefix("golaxy:mutex:").Apply(options)
		With.CustomAuth("", "").Apply(options)
		With.CustomAddress("127.0.0.1:6379").Apply(options)
		With.CustomDB(0).Apply(options)
	}
}

// RedisClient redis客户端（优先使用）
func (_RedisSyncOption) RedisClient(cli *redis.Client) option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		options.RedisClient = cli
	}
}

// RedisConfig redis配置（次优先使用）
func (_RedisSyncOption) RedisConfig(conf *redis.Options) option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		options.RedisConfig = conf
	}
}

// RedisURL redis连接url（次次优先使用）
func (_RedisSyncOption) RedisURL(url string) option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		options.RedisURL = url
	}
}

// KeyPrefix key前缀
func (_RedisSyncOption) KeyPrefix(prefix string) option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		options.KeyPrefix = prefix
	}
}

// CustomAuth 自定义redis认证信息（次次次优先使用）
func (_RedisSyncOption) CustomAuth(username, password string) option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddress 自定义redis连接地址信息（次次次优先使用）
func (_RedisSyncOption) CustomAddress(addr string) option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			exception.Panicf("dsync: %w: %w", core.ErrArgs, err)
		}
		options.CustomAddress = addr
	}
}

// CustomDB 自定义redis数据库id（次次次优先使用）
func (_RedisSyncOption) CustomDB(db int) option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		options.CustomDB = db
	}
}
