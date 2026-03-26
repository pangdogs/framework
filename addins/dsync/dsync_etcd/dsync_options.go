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

package dsync_etcd

import (
	"crypto/tls"
	"net"
	"strings"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdSyncOptions 所有选项
type EtcdSyncOptions struct {
	EtcdClient      *clientv3.Client
	EtcdConfig      *clientv3.Config
	KeyPrefix       string
	CustomUsername  string
	CustomPassword  string
	CustomAddresses []string
	CustomTLSConfig *tls.Config
}

var With _EtcdSyncOption

type _EtcdSyncOption struct{}

// Default 默认选项
func (_EtcdSyncOption) Default() option.Setting[EtcdSyncOptions] {
	return func(options *EtcdSyncOptions) {
		With.EtcdClient(nil).Apply(options)
		With.EtcdConfig(nil).Apply(options)
		With.KeyPrefix("/golaxy/mutex/").Apply(options)
		With.CustomAuth("", "").Apply(options)
		With.CustomAddresses("127.0.0.1:2379").Apply(options)
		With.CustomTLSConfig(nil).Apply(options)
	}
}

// EtcdClient etcd客户端（优先使用）
func (_EtcdSyncOption) EtcdClient(cli *clientv3.Client) option.Setting[EtcdSyncOptions] {
	return func(options *EtcdSyncOptions) {
		options.EtcdClient = cli
	}
}

// EtcdConfig etcd配置（次优先使用）
func (_EtcdSyncOption) EtcdConfig(config *clientv3.Config) option.Setting[EtcdSyncOptions] {
	return func(options *EtcdSyncOptions) {
		options.EtcdConfig = config
	}
}

// KeyPrefix key前缀
func (_EtcdSyncOption) KeyPrefix(prefix string) option.Setting[EtcdSyncOptions] {
	return func(options *EtcdSyncOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.KeyPrefix = prefix
	}
}

// CustomAuth 自定义etcd认证信息（次次优先使用）
func (_EtcdSyncOption) CustomAuth(username, password string) option.Setting[EtcdSyncOptions] {
	return func(options *EtcdSyncOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 自定义etcd连接地址信息（次次优先使用）
func (_EtcdSyncOption) CustomAddresses(addrs ...string) option.Setting[EtcdSyncOptions] {
	return func(options *EtcdSyncOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				exception.Panicf("dsync: %w: %w", core.ErrArgs, err)
			}
		}
		options.CustomAddresses = addrs
	}
}

// CustomTLSConfig 自定义etcd的TLS加密（次次优先使用）
func (_EtcdSyncOption) CustomTLSConfig(conf *tls.Config) option.Setting[EtcdSyncOptions] {
	return func(options *EtcdSyncOptions) {
		options.CustomTLSConfig = conf
	}
}
