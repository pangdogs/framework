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

// DistEntityQuerierOptions 所有选项
type DistEntityQuerierOptions struct {
	EtcdClient       *clientv3.Client
	EtcdConfig       *clientv3.Config
	KeyPrefix        string
	CacheNumCounters int64
	CacheMaxCost     int64
	CacheBufferItems int64
	CacheTTL         time.Duration
	CustomUsername   string
	CustomPassword   string
	CustomAddresses  []string
	CustomTLSConfig  *tls.Config
}

type _DistEntityQuerierOption struct{}

// Default 默认值
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

// EtcdClient etcd客户端，最优先使用
func (_DistEntityQuerierOption) EtcdClient(cli *clientv3.Client) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		options.EtcdClient = cli
	}
}

// EtcdConfig etcd配置，次优先使用
func (_DistEntityQuerierOption) EtcdConfig(config *clientv3.Config) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		options.EtcdConfig = config
	}
}

// KeyPrefix 所有key的前缀
func (_DistEntityQuerierOption) KeyPrefix(prefix string) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// CacheNumCounters 缓存LFU计数器数量
func (_DistEntityQuerierOption) CacheNumCounters(n int64) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		if n <= 0 {
			exception.Panicf("dent: %w: option CacheNumCounters can't be set to a value less equal 0", core.ErrArgs)
		}
		options.CacheNumCounters = n
	}
}

// CacheMaxCost 缓存容量限制，超过将触发LFU淘汰
func (_DistEntityQuerierOption) CacheMaxCost(n int64) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		if n <= 0 {
			exception.Panicf("dent: %w: option CacheMaxCost can't be set to a value less equal 0", core.ErrArgs)
		}
		options.CacheMaxCost = n
	}
}

// CacheBufferItems 缓存并发缓冲大小
func (_DistEntityQuerierOption) CacheBufferItems(n int64) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		if n <= 0 {
			exception.Panicf("dent: %w: option CacheBufferItems can't be set to a value less equal 0", core.ErrArgs)
		}
		options.CacheBufferItems = n
	}
}

// CacheTTL 缓存TTL
func (_DistEntityQuerierOption) CacheTTL(ttl time.Duration) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		if ttl < 3*time.Second {
			exception.Panicf("dent: %w: option CacheTTL can't be set to a value less than 3 seconds", core.ErrArgs)
		}
		options.CacheTTL = ttl
	}
}

// CustomAuth 自定义设置etcd鉴权信息
func (_DistEntityQuerierOption) CustomAuth(username, password string) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 自定义设置etcd服务地址
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

// CustomTLSConfig 自定义设置加密etcd连接的配置
func (_DistEntityQuerierOption) CustomTLSConfig(conf *tls.Config) option.Setting[DistEntityQuerierOptions] {
	return func(options *DistEntityQuerierOptions) {
		options.CustomTLSConfig = conf
	}
}
