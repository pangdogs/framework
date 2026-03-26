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

package discovery_etcd

import (
	"crypto/tls"
	"net"
	"strings"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// RegistryOptions 所有选项
type RegistryOptions struct {
	EtcdClient      *clientv3.Client
	EtcdConfig      *clientv3.Config
	KeyPrefix       string
	CustomUsername  string
	CustomPassword  string
	CustomAddresses []string
	CustomTLSConfig *tls.Config
}

var With _RegistryOption

type _RegistryOption struct{}

// Default 默认值
func (_RegistryOption) Default() option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		With.EtcdClient(nil).Apply(options)
		With.EtcdConfig(nil).Apply(options)
		With.KeyPrefix("/golaxy/svc/").Apply(options)
		With.CustomAuth("", "").Apply(options)
		With.CustomAddresses("127.0.0.1:2379").Apply(options)
		With.CustomTLSConfig(nil).Apply(options)
	}
}

// EtcdClient etcd客户端，最优先使用
func (_RegistryOption) EtcdClient(cli *clientv3.Client) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.EtcdClient = cli
	}
}

// EtcdConfig etcd配置，次优先使用
func (_RegistryOption) EtcdConfig(config *clientv3.Config) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.EtcdConfig = config
	}
}

// KeyPrefix 所有key的前缀
func (_RegistryOption) KeyPrefix(prefix string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// CustomAuth 自定义设置etcd鉴权信息
func (_RegistryOption) CustomAuth(username, password string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 自定义设置etcd服务地址
func (_RegistryOption) CustomAddresses(addrs ...string) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				exception.Panicf("registry: %w: %w", core.ErrArgs, err)
			}
		}
		options.CustomAddresses = addrs
	}
}

// CustomTLSConfig 自定义设置加密etcd连接的配置
func (_RegistryOption) CustomTLSConfig(conf *tls.Config) option.Setting[RegistryOptions] {
	return func(options *RegistryOptions) {
		options.CustomTLSConfig = conf
	}
}
