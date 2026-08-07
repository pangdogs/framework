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

// RedisSyncOptions 配置 Redis 分布式锁实现的客户端、连接及键空间。
type RedisSyncOptions struct {
	RedisClient    *redis.Client  // RedisClient 非 nil 时直接复用，停止时不会关闭它。
	RedisConfig    *redis.Options // RedisConfig 在未提供客户端时优先于 RedisURL 和 Custom 字段。
	RedisURL       string         // RedisURL 在未提供完整配置时用于解析连接选项。
	KeyPrefix      string         // KeyPrefix 是所有锁键的公共前缀。
	CustomUsername string         // CustomUsername 是自行构造配置时使用的用户名。
	CustomPassword string         // CustomPassword 是自行构造配置时使用的密码。
	CustomAddress  string         // CustomAddress 是自行构造配置时使用的服务地址。
	CustomDB       int            // CustomDB 是自行构造配置时使用的数据库编号。
}

// With 提供 Redis 分布式锁 add-in 的 Option 构造方法。
var With _RedisSyncOption

type _RedisSyncOption struct{}

// Default 返回本地 Redis 0 号库及 golaxy:mutex: 键前缀的默认设置。
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

// RedisClient 设置要复用的 Redis 客户端，其优先级最高。
func (_RedisSyncOption) RedisClient(cli *redis.Client) option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		options.RedisClient = cli
	}
}

// RedisConfig 设置创建 Redis 客户端时使用的完整配置，其优先级次于 RedisClient。
func (_RedisSyncOption) RedisConfig(conf *redis.Options) option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		options.RedisConfig = conf
	}
}

// RedisURL 设置 Redis 连接 URL，其优先级次于 RedisConfig。
func (_RedisSyncOption) RedisURL(url string) option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		options.RedisURL = url
	}
}

// KeyPrefix 设置锁键前缀；非空值会自动补充末尾冒号。
func (_RedisSyncOption) KeyPrefix(prefix string) option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ":") {
			prefix += ":"
		}
		options.KeyPrefix = prefix
	}
}

// CustomAuth 设置自行构造 Redis 配置时使用的用户名和密码。
func (_RedisSyncOption) CustomAuth(username, password string) option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddress 设置自行构造 Redis 配置时使用的地址，并校验 host:port 格式。
func (_RedisSyncOption) CustomAddress(addr string) option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			exception.Panicf("dsync: %w: %w", core.ErrArgs, err)
		}
		options.CustomAddress = addr
	}
}

// CustomDB 设置自行构造 Redis 配置时使用的数据库编号。
func (_RedisSyncOption) CustomDB(db int) option.Setting[RedisSyncOptions] {
	return func(options *RedisSyncOptions) {
		options.CustomDB = db
	}
}
