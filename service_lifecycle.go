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
	Built(inst IServiceInstance)
}

type LifecycleServiceBirth interface {
	Birth(inst IServiceInstance)
}

type LifecycleServiceStarting interface {
	Starting(inst IServiceInstance)
}

type LifecycleServiceStarted interface {
	Started(inst IServiceInstance)
}

type LifecycleServiceTerminating interface {
	Terminating(inst IServiceInstance)
}

type LifecycleServiceTerminated interface {
	Terminated(inst IServiceInstance)
}

type LifecycleServiceAddInActivating interface {
	AddInActivating(inst IServiceInstance, addIn extension.AddInStatus)
}

type LifecycleServiceAddInActivated interface {
	AddInActivated(inst IServiceInstance, addIn extension.AddInStatus)
}

type LifecycleServiceAddInDeactivating interface {
	AddInDeactivating(inst IServiceInstance, addIn extension.AddInStatus)
}

type LifecycleServiceAddInDeactivated interface {
	AddInDeactivated(inst IServiceInstance, addIn extension.AddInStatus)
}

type LifecycleServiceEntityPTDeclared interface {
	EntityPTDeclared(inst IServiceInstance, entityPT ec.EntityPT)
}

type LifecycleServiceEntityPTRedeclared interface {
	EntityPTRedeclared(inst IServiceInstance, entityPT ec.EntityPT)
}

type LifecycleServiceEntityPTUndeclared interface {
	EntityPTUndeclared(inst IServiceInstance, entityPT ec.EntityPT)
}
