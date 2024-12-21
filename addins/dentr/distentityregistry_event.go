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

//go:generate go run git.golaxy.org/core/event/eventc event --default_export=false
//go:generate go run git.golaxy.org/core/event/eventc eventtab --name=distEntityRegistryEventTab
package dentr

import (
	"git.golaxy.org/core/ec"
)

// EventDistEntityOnline 事件：分布式实体上线
type EventDistEntityOnline interface {
	OnDistEntityOnline(entity ec.Entity)
}

// EventDistEntityOffline 事件：分布式实体下线
type EventDistEntityOffline interface {
	OnDistEntityOffline(entity ec.Entity)
}
