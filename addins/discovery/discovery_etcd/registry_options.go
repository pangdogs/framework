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

// EtcdRegistryOptions 配置服务发现实现使用的 ETCD 客户端与键空间。
type EtcdRegistryOptions struct {
	EtcdClient      *clientv3.Client // EtcdClient 非 nil 时直接复用，停止时不会关闭它。
	EtcdConfig      *clientv3.Config // EtcdConfig 在未提供客户端时优先于 Custom 字段。
	KeyPrefix       string           // KeyPrefix 是所有服务注册键的公共前缀。
	CustomUsername  string           // CustomUsername 是自行构造客户端时使用的用户名。
	CustomPassword  string           // CustomPassword 是自行构造客户端时使用的密码。
	CustomAddresses []string         // CustomAddresses 是自行构造客户端时使用的端点。
	CustomTLSConfig *tls.Config      // CustomTLSConfig 是自行构造客户端时使用的 TLS 配置。
}

// With 提供 ETCD 服务发现 add-in 的 Option 构造方法。
var With _EtcdRegistryOption

type _EtcdRegistryOption struct{}

// Default 返回使用本地 ETCD 端点及 /golaxy/svc/ 键前缀的默认设置。
func (_EtcdRegistryOption) Default() option.Setting[EtcdRegistryOptions] {
	return func(options *EtcdRegistryOptions) {
		With.EtcdClient(nil).Apply(options)
		With.EtcdConfig(nil).Apply(options)
		With.KeyPrefix("/golaxy/svc/").Apply(options)
		With.CustomAuth("", "").Apply(options)
		With.CustomAddresses("127.0.0.1:2379").Apply(options)
		With.CustomTLSConfig(nil).Apply(options)
	}
}

// EtcdClient 设置要复用的 ETCD 客户端，其优先级最高。
func (_EtcdRegistryOption) EtcdClient(cli *clientv3.Client) option.Setting[EtcdRegistryOptions] {
	return func(options *EtcdRegistryOptions) {
		options.EtcdClient = cli
	}
}

// EtcdConfig 设置创建 ETCD 客户端时使用的完整配置，其优先级次于 EtcdClient。
func (_EtcdRegistryOption) EtcdConfig(config *clientv3.Config) option.Setting[EtcdRegistryOptions] {
	return func(options *EtcdRegistryOptions) {
		options.EtcdConfig = config
	}
}

// KeyPrefix 设置服务注册键前缀；非空值会自动补充末尾斜杠。
func (_EtcdRegistryOption) KeyPrefix(prefix string) option.Setting[EtcdRegistryOptions] {
	return func(options *EtcdRegistryOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// CustomAuth 设置自行构造 ETCD 客户端时使用的用户名和密码。
func (_EtcdRegistryOption) CustomAuth(username, password string) option.Setting[EtcdRegistryOptions] {
	return func(options *EtcdRegistryOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 设置自行构造 ETCD 客户端时使用的端点，并校验 host:port 格式。
func (_EtcdRegistryOption) CustomAddresses(addrs ...string) option.Setting[EtcdRegistryOptions] {
	return func(options *EtcdRegistryOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				exception.Panicf("registry: %w: %w", core.ErrArgs, err)
			}
		}
		options.CustomAddresses = addrs
	}
}

// CustomTLSConfig 设置自行构造 ETCD 客户端时使用的 TLS 配置。
func (_EtcdRegistryOption) CustomTLSConfig(conf *tls.Config) option.Setting[EtcdRegistryOptions] {
	return func(options *EtcdRegistryOptions) {
		options.CustomTLSConfig = conf
	}
}
