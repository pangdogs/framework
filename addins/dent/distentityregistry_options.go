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

// DistEntityRegistryOptions 所有选项
type DistEntityRegistryOptions struct {
	EtcdClient      *clientv3.Client
	EtcdConfig      *clientv3.Config
	KeyPrefix       string
	RegistrationTTL time.Duration
	CustomUsername  string
	CustomPassword  string
	CustomAddresses []string
	CustomTLSConfig *tls.Config
}

type _DistEntityRegistryOption struct{}

// Default 默认值
func (_DistEntityRegistryOption) Default() option.Setting[DistEntityRegistryOptions] {
	return func(options *DistEntityRegistryOptions) {
		With.Registry.EtcdClient(nil).Apply(options)
		With.Registry.EtcdConfig(nil).Apply(options)
		With.Registry.KeyPrefix("/golaxy/dent/").Apply(options)
		With.Registry.RegistrationTTL(time.Minute).Apply(options)
		With.Registry.CustomAuth("", "").Apply(options)
		With.Registry.CustomAddresses("127.0.0.1:2379").Apply(options)
		With.Registry.CustomTLSConfig(nil).Apply(options)
	}
}

// EtcdClient etcd客户端，最优先使用
func (_DistEntityRegistryOption) EtcdClient(cli *clientv3.Client) option.Setting[DistEntityRegistryOptions] {
	return func(options *DistEntityRegistryOptions) {
		options.EtcdClient = cli
	}
}

// EtcdConfig etcd配置，次优先使用
func (_DistEntityRegistryOption) EtcdConfig(config *clientv3.Config) option.Setting[DistEntityRegistryOptions] {
	return func(options *DistEntityRegistryOptions) {
		options.EtcdConfig = config
	}
}

// KeyPrefix 所有key的前缀
func (_DistEntityRegistryOption) KeyPrefix(prefix string) option.Setting[DistEntityRegistryOptions] {
	return func(options *DistEntityRegistryOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// RegistrationTTL 注册实体信息TTL
func (_DistEntityRegistryOption) RegistrationTTL(ttl time.Duration) option.Setting[DistEntityRegistryOptions] {
	return func(options *DistEntityRegistryOptions) {
		if ttl < 3*time.Second {
			exception.Panicf("dent: %w: option RegistrationTTL can't be set to a value less than 3 seconds", core.ErrArgs)
		}
		options.RegistrationTTL = ttl
	}
}

// CustomAuth 自定义设置etcd鉴权信息
func (_DistEntityRegistryOption) CustomAuth(username, password string) option.Setting[DistEntityRegistryOptions] {
	return func(options *DistEntityRegistryOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 自定义设置etcd服务地址
func (_DistEntityRegistryOption) CustomAddresses(addrs ...string) option.Setting[DistEntityRegistryOptions] {
	return func(options *DistEntityRegistryOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				exception.Panicf("dent: %w: %w", core.ErrArgs, err)
			}
		}
		options.CustomAddresses = addrs
	}
}

// CustomTLSConfig 自定义设置加密etcd连接的配置
func (_DistEntityRegistryOption) CustomTLSConfig(conf *tls.Config) option.Setting[DistEntityRegistryOptions] {
	return func(options *DistEntityRegistryOptions) {
		options.CustomTLSConfig = conf
	}
}
