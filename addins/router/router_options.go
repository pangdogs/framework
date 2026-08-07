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

package router

import (
	"crypto/tls"
	"net"
	"strings"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// RouterOptions 配置路由组使用的 ETCD 客户端及键空间。
type RouterOptions struct {
	EtcdClient      *clientv3.Client // EtcdClient 非 nil 时直接复用，停止时不会关闭它。
	EtcdConfig      *clientv3.Config // EtcdConfig 在未提供客户端时优先于 Custom 字段。
	GroupKeyPrefix  string           // GroupKeyPrefix 是路由组记录的键前缀。
	EntityKeyPrefix string           // EntityKeyPrefix 是实体反向索引的键前缀。
	CustomUsername  string           // CustomUsername 是自行构造客户端时使用的用户名。
	CustomPassword  string           // CustomPassword 是自行构造客户端时使用的密码。
	CustomAddresses []string         // CustomAddresses 是自行构造客户端时使用的端点。
	CustomTLSConfig *tls.Config      // CustomTLSConfig 是自行构造客户端时使用的 TLS 配置。
}

// With 提供路由 add-in 的 Option 构造方法。
var With _RouterOption

type _RouterOption struct{}

// Default 返回使用本地 ETCD 端点及默认 group/entity 键前缀的设置。
func (_RouterOption) Default() option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		With.EtcdClient(nil)(options)
		With.EtcdConfig(nil)(options)
		With.GroupKeyPrefix("/golaxy/router/group/")(options)
		With.EntityKeyPrefix("/golaxy/router/entity/")(options)
		With.CustomAuth("", "")(options)
		With.CustomAddresses("127.0.0.1:2379")(options)
		With.CustomTLSConfig(nil)(options)
	}
}

// EtcdClient 设置要复用的 ETCD 客户端，其优先级最高。
func (_RouterOption) EtcdClient(cli *clientv3.Client) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		options.EtcdClient = cli
	}
}

// EtcdConfig 设置创建 ETCD 客户端时使用的完整配置，其优先级次于 EtcdClient。
func (_RouterOption) EtcdConfig(config *clientv3.Config) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		options.EtcdConfig = config
	}
}

// GroupKeyPrefix 设置路由组记录的键前缀；非空值会自动补充末尾斜杠。
func (_RouterOption) GroupKeyPrefix(prefix string) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.GroupKeyPrefix = prefix
	}
}

// EntityKeyPrefix 设置实体反向索引的键前缀；非空值会自动补充末尾斜杠。
func (_RouterOption) EntityKeyPrefix(prefix string) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.EntityKeyPrefix = prefix
	}
}

// CustomAuth 设置自行构造 ETCD 客户端时使用的用户名和密码。
func (_RouterOption) CustomAuth(username, password string) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 设置自行构造 ETCD 客户端时使用的端点，并校验 host:port 格式。
func (_RouterOption) CustomAddresses(addrs ...string) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				exception.Panicf("router: %w: %w", core.ErrArgs, err)
			}
		}
		options.CustomAddresses = addrs
	}
}

// CustomTLSConfig 设置自行构造 ETCD 客户端时使用的 TLS 配置。
func (_RouterOption) CustomTLSConfig(conf *tls.Config) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		options.CustomTLSConfig = conf
	}
}
