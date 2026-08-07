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

// DistEntityRegistryOptions 配置分布式实体注册端的 ETCD 连接与租约。
type DistEntityRegistryOptions struct {
	EtcdClient      *clientv3.Client // EtcdClient 非 nil 时直接复用，停止时不会关闭它。
	EtcdConfig      *clientv3.Config // EtcdConfig 在未提供客户端时优先于 Custom 字段。
	KeyPrefix       string           // KeyPrefix 是所有实体注册键的公共前缀。
	RegistrationTTL time.Duration    // RegistrationTTL 是实体注册租约的有效期。
	CustomUsername  string           // CustomUsername 是自行构造客户端时使用的用户名。
	CustomPassword  string           // CustomPassword 是自行构造客户端时使用的密码。
	CustomAddresses []string         // CustomAddresses 是自行构造客户端时使用的端点。
	CustomTLSConfig *tls.Config      // CustomTLSConfig 是自行构造客户端时使用的 TLS 配置。
}

type _DistEntityRegistryOption struct{}

// Default 返回使用本地 ETCD 端点、/golaxy/dent/ 键前缀和一分钟租约的默认设置。
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

// EtcdClient 设置要复用的 ETCD 客户端，其优先级最高。
func (_DistEntityRegistryOption) EtcdClient(cli *clientv3.Client) option.Setting[DistEntityRegistryOptions] {
	return func(options *DistEntityRegistryOptions) {
		options.EtcdClient = cli
	}
}

// EtcdConfig 设置创建 ETCD 客户端时使用的完整配置，其优先级次于 EtcdClient。
func (_DistEntityRegistryOption) EtcdConfig(config *clientv3.Config) option.Setting[DistEntityRegistryOptions] {
	return func(options *DistEntityRegistryOptions) {
		options.EtcdConfig = config
	}
}

// KeyPrefix 设置实体注册键前缀；非空值会自动补充末尾斜杠。
func (_DistEntityRegistryOption) KeyPrefix(prefix string) option.Setting[DistEntityRegistryOptions] {
	return func(options *DistEntityRegistryOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// RegistrationTTL 设置实体注册租约有效期，必须不少于三秒。
func (_DistEntityRegistryOption) RegistrationTTL(ttl time.Duration) option.Setting[DistEntityRegistryOptions] {
	return func(options *DistEntityRegistryOptions) {
		if ttl < 3*time.Second {
			exception.Panicf("dent: %w: option RegistrationTTL must be >= 3 seconds", core.ErrArgs)
		}
		options.RegistrationTTL = ttl
	}
}

// CustomAuth 设置自行构造 ETCD 客户端时使用的用户名和密码。
func (_DistEntityRegistryOption) CustomAuth(username, password string) option.Setting[DistEntityRegistryOptions] {
	return func(options *DistEntityRegistryOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 设置自行构造 ETCD 客户端时使用的端点，并校验 host:port 格式。
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

// CustomTLSConfig 设置自行构造 ETCD 客户端时使用的 TLS 配置。
func (_DistEntityRegistryOption) CustomTLSConfig(conf *tls.Config) option.Setting[DistEntityRegistryOptions] {
	return func(options *DistEntityRegistryOptions) {
		options.CustomTLSConfig = conf
	}
}
