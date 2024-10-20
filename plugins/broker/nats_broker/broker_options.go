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

package nats_broker

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
	"github.com/nats-io/nats.go"
	"net"
	"strings"
)

// BrokerOptions is a struct that holds various configuration options for the NATS broker.
type BrokerOptions struct {
	NatsClient      *nats.Conn
	TopicPrefix     string
	QueuePrefix     string
	CustomAddresses []string
	CustomUsername  string
	CustomPassword  string
}

var With _Option

type _Option struct{}

// Default sets default values for BrokerOptions.
func (_Option) Default() option.Setting[BrokerOptions] {
	return func(options *BrokerOptions) {
		With.NatsClient(nil)(options)
		With.TopicPrefix("")(options)
		With.QueuePrefix("")(options)
		With.CustomAuth("", "")(options)
		With.CustomAddresses("127.0.0.1:4222")(options)
	}
}

// NatsClient sets the NATS client in BrokerOptions.
func (_Option) NatsClient(cli *nats.Conn) option.Setting[BrokerOptions] {
	return func(o *BrokerOptions) {
		o.NatsClient = cli
	}
}

// TopicPrefix sets the topic prefix in BrokerOptions.
func (_Option) TopicPrefix(prefix string) option.Setting[BrokerOptions] {
	return func(o *BrokerOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
		o.TopicPrefix = prefix
	}
}

// QueuePrefix sets the queue prefix in BrokerOptions.
func (_Option) QueuePrefix(prefix string) option.Setting[BrokerOptions] {
	return func(o *BrokerOptions) {
		if prefix != "" && !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
		o.QueuePrefix = prefix
	}
}

// CustomAuth sets the authentication credentials in BrokerOptions. If NatsClient is nil, these credentials are used for authentication.
func (_Option) CustomAuth(username, password string) option.Setting[BrokerOptions] {
	return func(options *BrokerOptions) {
		options.CustomUsername = username
		options.CustomPassword = password
	}
}

// CustomAddresses sets the addresses in BrokerOptions. If NatsClient is nil, these addresses are used as the connection addresses.
func (_Option) CustomAddresses(addrs ...string) option.Setting[BrokerOptions] {
	return func(options *BrokerOptions) {
		for _, addr := range addrs {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				exception.Panicf("%w: %w", core.ErrArgs, err)
			}
		}
		options.CustomAddresses = addrs
	}
}
