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

type LifecycleServiceBirth interface {
	OnBirth(svc IService)
}

type LifecycleServiceBuilt interface {
	OnBuilt(svc IService)
}

type LifecycleServiceStarting interface {
	OnStarting(svc IService)
}

type LifecycleServiceStarted interface {
	OnStarted(svc IService)
}

type LifecycleServiceTerminating interface {
	OnTerminating(svc IService)
}

type LifecycleServiceTerminated interface {
	OnTerminated(svc IService)
}

type LifecycleServiceAddInActivating interface {
	OnAddInActivating(svc IService, addIn extension.AddInStatus)
}

type LifecycleServiceAddInActivated interface {
	OnAddInActivated(svc IService, addIn extension.AddInStatus)
}

type LifecycleServiceAddInDeactivating interface {
	OnAddInDeactivating(svc IService, addIn extension.AddInStatus)
}

type LifecycleServiceAddInDeactivated interface {
	OnAddInDeactivated(svc IService, addIn extension.AddInStatus)
}

type LifecycleServiceEntityPTDeclared interface {
	OnEntityPTDeclared(svc IService, entityPT ec.EntityPT)
}

type LifecycleServiceComponentPTDeclared interface {
	OnComponentPTDeclared(svc IService, componentPT ec.ComponentPT)
}

type LifecycleServiceEntityRegistered interface {
	OnEntityRegistered(svc IService, entity ec.ConcurrentEntity)
}

type LifecycleServiceEntityDeregistered interface {
	OnEntityDeregistered(svc IService, entity ec.ConcurrentEntity)
}
