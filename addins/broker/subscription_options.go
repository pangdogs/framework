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

package broker

import (
	"git.golaxy.org/core/utils/option"
)

// SubscribeOptions 所有订阅选项
type SubscribeOptions struct {
	// AutoAck 自动确认
	AutoAck bool
	// Queue 订阅队列组
	Queue string
}

var With _SubscribeOption

type _SubscribeOption struct{}

// Default 默认订阅选项
func (_SubscribeOption) Default() option.Setting[SubscribeOptions] {
	return func(options *SubscribeOptions) {
		With.AutoAck(true).Apply(options)
		With.Queue("").Apply(options)
	}
}

// AutoAck 自动确认
func (_SubscribeOption) AutoAck(b bool) option.Setting[SubscribeOptions] {
	return func(options *SubscribeOptions) {
		options.AutoAck = b
	}
}

// Queue 订阅队列组
func (_SubscribeOption) Queue(queue string) option.Setting[SubscribeOptions] {
	return func(options *SubscribeOptions) {
		options.Queue = queue
	}
}
