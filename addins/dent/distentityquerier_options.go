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

package dent

import (
	"crypto/tls"
	"net"
	"strings"
	"time"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// DistEntityQuerierOptions 配置分布式实体查询端的 ETCD 连接与本地缓存。
type DistEntityQuerierOptions struct {
	EtcdClient       *clientv3.Client // EtcdClient 非 nil 时直接复用，停止时不会关闭它。
	EtcdConfig       *clientv3.Config // EtcdConfig 在未提供客户端时优先于 Custom 字段。
	KeyPrefix        string           // KeyPrefix 是所有实体注册键的公共前缀。
	CacheNumCounters int64            // CacheNumCounters 是 Ristretto 访问计数器数量。
	CacheMaxCost     int64            // CacheMaxCost 是缓存总成本上限；每项成本为 1。
	CacheBufferItems int64            // CacheBufferItems 是 Ristretto 写缓冲区大小。
	CacheTTL         time.Duration    // CacheTTL 是查询结果的最长缓存时间。
	CustomUsername   string           // CustomUsername 是自行构造客户端时使用的用户名。
	CustomPassword   string           // CustomPassword 是自行构造客户端时使用的密码。
	CustomAddresses  []string         // CustomAddresses 是自行构造客户端时使用的端点。
	CustomTLSConfig  *tls.Config      // CustomTLSConfig 是自行构造客户端时使用的 TLS 配置。
}

type _DistEntityQuerierOption struct{}

// Default 返回使用本地 ETCD 端点、/golaxy/dent/ 键前缀和十分钟缓存的默认设置。
func (_DistEntityQuerierOption) Default() option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		With.Querier.EtcdClient(nil).Apply(options)
		With.Querier.EtcdConfig(nil).Apply(options)
		With.Querier.KeyPrefix("/golaxy/dent/").Apply(options)
		With.Querier.CacheNumCounters(100000).Apply(options)
		With.Querier.CacheMaxCost(100000).Apply(options)
		With.Querier.CacheBufferItems(128).Apply(options)
		With.Querier.CacheTTL(10 * time.Minute).Apply(options)
		With.Querier.CustomAuth("", "").Apply(options)
		With.Querier.CustomAddresses("127.0.0.1:2379").Apply(options)
		With.Querier.CustomTLSConfig(nil).Apply(options)
	}
}

// EtcdClient 设置要复用的 ETCD 客户端，其优先级最高。
func (_DistEntityQuerierOption) EtcdClient(cli *clientv3.Client) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		options.EtcdClient = cli
	}
}

// EtcdConfig 设置创建 ETCD 客户端时使用的完整配置，其优先级次于 EtcdClient。
func (_DistEntityQuerierOption) EtcdConfig(config *clientv3.Config) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		options.EtcdConfig = config
	}
}

// KeyPrefix 设置实体注册键前缀；非空值会自动补充末尾斜杠。
func (_DistEntityQuerierOption) KeyPrefix(prefix string) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// CacheNumCounters 设置缓存 LFU 计数器数量，必须大于 0。
func (_DistEntityQuerierOption) CacheNumCounters(n int64) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		if n <= 0 {
			exception.Panicf("dent: %w: option CacheNumCounters must be > 0", core.ErrArgs)
		}
		options.CacheNumCounters = n
	}
}

// CacheMaxCost 设置缓存成本上限，超过后触发 LFU 淘汰；必须大于 0。
func (_DistEntityQuerierOption) CacheMaxCost(n int64) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		if n <= 0 {
			exception.Panicf("dent: %w: option CacheMaxCost must be > 0", core.ErrArgs)
		}
		options.CacheMaxCost = n
	}
}

// CacheBufferItems 设置缓存写缓冲区大小，必须大于 0。
func (_DistEntityQuerierOption) CacheBufferItems(n int64) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		if n <= 0 {
			exception.Panicf("dent: %w: option CacheBufferItems must be > 0", core.ErrArgs)
		}
		options.CacheBufferItems = n
	}
}

// CacheTTL 设置查询结果缓存时间，必须不少于三秒。
func (_DistEntityQuerierOption) CacheTTL(ttl time.Duration) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		if ttl < 3*time.Second {
			exception.Panicf("dent: %w: option CacheTTL must be >= 3 seconds", core.ErrArgs)
		}
		options.CacheTTL = ttl
	}
}

// CustomAuth 设置自行构造 ETCD 客户端时使用的用户名和密码。
func (_DistEntityQuerierOption) CustomAuth(username, password string) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 设置自行构造 ETCD 客户端时使用的端点，并校验 host:port 格式。
func (_DistEntityQuerierOption) CustomAddresses(addrs ...string) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				exception.Panicf("dentq: %w: %w", core.ErrArgs, err)
			}
		}
		options.CustomAddresses = addrs
	}
}

// CustomTLSConfig 设置自行构造 ETCD 客户端时使用的 TLS 配置。
func (_DistEntityQuerierOption) CustomTLSConfig(conf *tls.Config) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		options.CustomTLSConfig = conf
	}
}
