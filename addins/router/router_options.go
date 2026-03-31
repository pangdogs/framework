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

type RouterOptions struct {
	EtcdClient      *clientv3.Client
	EtcdConfig      *clientv3.Config
	GroupKeyPrefix  string
	EntityKeyPrefix string
	CustomUsername  string
	CustomPassword  string
	CustomAddresses []string
	CustomTLSConfig *tls.Config
}

var With _RouterOption

type _RouterOption struct{}

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

func (_RouterOption) EtcdClient(cli *clientv3.Client) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		options.EtcdClient = cli
	}
}

func (_RouterOption) EtcdConfig(config *clientv3.Config) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		options.EtcdConfig = config
	}
}

func (_RouterOption) GroupKeyPrefix(prefix string) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.GroupKeyPrefix = prefix
	}
}

func (_RouterOption) EntityKeyPrefix(prefix string) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		options.EntityKeyPrefix = prefix
	}
}

func (_RouterOption) CustomAuth(username, password string) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

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

func (_RouterOption) CustomTLSConfig(conf *tls.Config) option.Setting[RouterOptions] {
	return func(options *RouterOptions) {
		options.CustomTLSConfig = conf
	}
}
