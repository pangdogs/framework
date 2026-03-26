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

package broker_nats

import (
	"net"
	"strings"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
	"github.com/nats-io/nats.go"
)

// NatsBrokerOptions 所有选项
type NatsBrokerOptions struct {
	NatsClient      *nats.Conn
	TopicPrefix     string
	QueuePrefix     string
	CustomAddresses []string
	CustomUsername  string
	CustomPassword  string
}

var With _NatsBrokerOption

type _NatsBrokerOption struct{}

// Default 默认选项
func (_NatsBrokerOption) Default() option.Setting[NatsBrokerOptions] {
	return func(options *NatsBrokerOptions) {
		With.NatsClient(nil).Apply(options)
		With.TopicPrefix("").Apply(options)
		With.QueuePrefix("").Apply(options)
		With.CustomAuth("", "").Apply(options)
		With.CustomAddresses("127.0.0.1:4222").Apply(options)
	}
}

// NatsClient nats客户端（优先使用）
func (_NatsBrokerOption) NatsClient(cli *nats.Conn) option.Setting[NatsBrokerOptions] {
	return func(options *NatsBrokerOptions) {
		options.NatsClient = cli
	}
}

// TopicPrefix 订阅话题前缀
func (_NatsBrokerOption) TopicPrefix(prefix string) option.Setting[NatsBrokerOptions] {
	return func(options *NatsBrokerOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
		options.TopicPrefix = prefix
	}
}

// QueuePrefix 订阅队列组前缀
func (_NatsBrokerOption) QueuePrefix(prefix string) option.Setting[NatsBrokerOptions] {
	return func(options *NatsBrokerOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
		options.QueuePrefix = prefix
	}
}

// CustomAuth 自定义认证信息
func (_NatsBrokerOption) CustomAuth(username, password string) option.Setting[NatsBrokerOptions] {
	return func(options *NatsBrokerOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 自定义地址
func (_NatsBrokerOption) CustomAddresses(addrs ...string) option.Setting[NatsBrokerOptions] {
	return func(options *NatsBrokerOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				exception.Panicf("broker: %w: %w", core.ErrArgs, err)
			}
		}
		options.CustomAddresses = addrs
	}
}
