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

package etcd_discovery

import (
	"crypto/tls"
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"strings"
	"time"
)

// RegistryOptions 所有选项
type RegistryOptions struct {
	EtcdClient      *clientv3.Client
	EtcdConfig      *clientv3.Config
	KeyPrefix       string
	WatchChanSize   int
	TTL             time.Duration
	AutoRefreshTTL  bool
	CustomUsername  string
	CustomPassword  string
	CustomAddresses []string
	CustomTLSConfig *tls.Config
}

var With _Option

type _Option struct{}

// Default 默认值
func (_Option) Default() option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		With.EtcdClient(nil)(options)
		With.EtcdConfig(nil)(options)
		With.KeyPrefix("/golaxy/services/")(options)
		With.WatchChanSize(128)(options)
		With.TTL(10*time.Second, true)(options)
		With.CustomAuth("", "")(options)
		With.CustomAddresses("127.0.0.1:2379")(options)
		With.CustomTLSConfig(nil)(options)
	}
}

// EtcdClient etcd客户端，最优先使用
func (_Option) EtcdClient(cli *clientv3.Client) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.EtcdClient = cli
	}
}

// EtcdConfig etcd配置，次优先使用
func (_Option) EtcdConfig(config *clientv3.Config) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.EtcdConfig = config
	}
}

// KeyPrefix 所有key的前缀
func (_Option) KeyPrefix(prefix string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// WatchChanSize 监控服务变化的channel大小
func (_Option) WatchChanSize(size int) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		if size < 0 {
			exception.Panicf("%w: option WatchChanSize can't be set to a value less than 0", core.ErrArgs)
		}
		options.WatchChanSize = size
	}
}

// TTL 默认TTL
func (_Option) TTL(ttl time.Duration, auto bool) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		if ttl < 3*time.Second {
			exception.Panicf("%w: option TTL can't be set to a value less than 3 second", core.ErrArgs)
		}
		options.TTL = ttl
		options.AutoRefreshTTL = auto
	}
}

// CustomAuth 自定义设置etcd鉴权信息
func (_Option) CustomAuth(username, password string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 自定义设置etcd服务地址
func (_Option) CustomAddresses(addrs ...string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				exception.Panicf("%w: %w", core.ErrArgs, err)
			}
		}
		options.CustomAddresses = addrs
	}
}

// CustomTLSConfig 自定义设置加密etcd连接的配置
func (_Option) CustomTLSConfig(conf *tls.Config) option.Setting[RegistryOptions] {
	return func(o *RegistryOptions) {
		o.CustomTLSConfig = conf
	}
}
