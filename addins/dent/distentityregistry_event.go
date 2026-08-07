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

//go:generate go run git.golaxy.org/core/event/eventc event --default_export_emit=false
//go:generate go run git.golaxy.org/core/event/eventc eventtab --name=distEntityRegistryEventTab
package dent

import (
	"git.golaxy.org/core/ec"
)

// EventDistEntityOnline 在全局实体完成分布式注册后同步通知监听者。
// +event-tab-gen:recursion=allow
type EventDistEntityOnline interface {
	// OnDistEntityOnline 处理已上线的全局实体。
	OnDistEntityOnline(entity ec.Entity)
}

// EventDistEntityOffline 在全局实体从当前运行时移除时同步通知监听者。
// +event-tab-gen:recursion=allow
type EventDistEntityOffline interface {
	// OnDistEntityOffline 处理已下线的全局实体。
	OnDistEntityOffline(entity ec.Entity)
}
