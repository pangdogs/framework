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
	Built(svc IService)
}

type LifecycleServiceBirth interface {
	Birth(svc IService)
}

type LifecycleServiceStarting interface {
	Starting(svc IService)
}

type LifecycleServiceStarted interface {
	Started(svc IService)
}

type LifecycleServiceTerminating interface {
	Terminating(svc IService)
}

type LifecycleServiceTerminated interface {
	Terminated(svc IService)
}

type LifecycleServiceAddInActivating interface {
	AddInActivating(svc IService, addIn extension.AddInStatus)
}

type LifecycleServiceAddInActivated interface {
	AddInActivated(svc IService, addIn extension.AddInStatus)
}

type LifecycleServiceAddInDeactivating interface {
	AddInDeactivating(svc IService, addIn extension.AddInStatus)
}

type LifecycleServiceAddInDeactivated interface {
	AddInDeactivated(svc IService, addIn extension.AddInStatus)
}

type LifecycleServiceEntityPTDeclared interface {
	EntityPTDeclared(svc IService, entityPT ec.EntityPT)
}

type LifecycleServiceEntityPTRedeclared interface {
	EntityPTRedeclared(svc IService, entityPT ec.EntityPT)
}

type LifecycleServiceEntityPTUndeclared interface {
	EntityPTUndeclared(svc IService, entityPT ec.EntityPT)
}
