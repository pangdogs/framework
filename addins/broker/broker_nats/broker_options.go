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

// NatsBrokerOptions 配置 NATS 连接及对外话题、队列组前缀。
type NatsBrokerOptions struct {
	NatsClient      *nats.Conn // NatsClient 非 nil 时直接复用，add-in 停止时不会关闭它。
	TopicPrefix     string     // TopicPrefix 会添加到发布和订阅话题前。
	QueuePrefix     string     // QueuePrefix 会添加到非空队列组名称前。
	CustomAddresses []string   // CustomAddresses 是自行建立连接时使用的服务地址。
	CustomUsername  string     // CustomUsername 是自行建立连接时使用的用户名。
	CustomPassword  string     // CustomPassword 是自行建立连接时使用的密码。
}

// With 提供 NATS broker 的 Option 构造方法。
var With _NatsBrokerOption

type _NatsBrokerOption struct{}

// Default 返回默认设置：连接 127.0.0.1:4222，且不添加话题或队列组前缀。
func (_NatsBrokerOption) Default() option.Setting[NatsBrokerOptions] {
	return func(options *NatsBrokerOptions) {
		With.NatsClient(nil).Apply(options)
		With.TopicPrefix("").Apply(options)
		With.QueuePrefix("").Apply(options)
		With.CustomAuth("", "").Apply(options)
		With.CustomAddresses("127.0.0.1:4222").Apply(options)
	}
}

// NatsClient 设置要复用的 NATS 客户端；非 nil 时忽略自定义连接参数。
func (_NatsBrokerOption) NatsClient(cli *nats.Conn) option.Setting[NatsBrokerOptions] {
	return func(options *NatsBrokerOptions) {
		options.NatsClient = cli
	}
}

// TopicPrefix 设置话题前缀；非空值会自动补充末尾的点号。
func (_NatsBrokerOption) TopicPrefix(prefix string) option.Setting[NatsBrokerOptions] {
	return func(options *NatsBrokerOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
		options.TopicPrefix = prefix
	}
}

// QueuePrefix 设置队列组前缀；非空值会自动补充末尾的点号。
func (_NatsBrokerOption) QueuePrefix(prefix string) option.Setting[NatsBrokerOptions] {
	return func(options *NatsBrokerOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
		options.QueuePrefix = prefix
	}
}

// CustomAuth 设置自行建立 NATS 连接时使用的用户名和密码。
func (_NatsBrokerOption) CustomAuth(username, password string) option.Setting[NatsBrokerOptions] {
	return func(options *NatsBrokerOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses 设置自行建立 NATS 连接时使用的服务地址，并校验 host:port 格式。
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
