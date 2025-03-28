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

package framework

import (
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/extension"
)

type LifecycleServiceBuilt interface {
	Built(inst IService)
}

type LifecycleServiceBirth interface {
	Birth(inst IService)
}

type LifecycleServiceStarting interface {
	Starting(inst IService)
}

type LifecycleServiceStarted interface {
	Started(inst IService)
}

type LifecycleServiceTerminating interface {
	Terminating(inst IService)
}

type LifecycleServiceTerminated interface {
	Terminated(inst IService)
}

type LifecycleServiceAddInActivating interface {
	AddInActivating(inst IService, addIn extension.AddInStatus)
}

type LifecycleServiceAddInActivated interface {
	AddInActivated(inst IService, addIn extension.AddInStatus)
}

type LifecycleServiceAddInDeactivating interface {
	AddInDeactivating(inst IService, addIn extension.AddInStatus)
}

type LifecycleServiceAddInDeactivated interface {
	AddInDeactivated(inst IService, addIn extension.AddInStatus)
}

type LifecycleServiceEntityPTDeclared interface {
	EntityPTDeclared(inst IService, entityPT ec.EntityPT)
}

type LifecycleServiceEntityPTRedeclared interface {
	EntityPTRedeclared(inst IService, entityPT ec.EntityPT)
}

type LifecycleServiceEntityPTUndeclared interface {
	EntityPTUndeclared(inst IService, entityPT ec.EntityPT)
}
