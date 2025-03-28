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

type InstallServiceLogger interface {
	InstallLogger(inst IService)
}

type InstallServiceConfig interface {
	InstallConfig(inst IService)
}

type InstallServiceBroker interface {
	InstallBroker(inst IService)
}

type InstallServiceRegistry interface {
	InstallRegistry(inst IService)
}

type InstallServiceDistSync interface {
	InstallDistSync(inst IService)
}

type InstallServiceDistService interface {
	InstallDistService(inst IService)
}

type InstallServiceRPC interface {
	InstallRPC(inst IService)
}

type InstallServiceDistEntityQuerier interface {
	InstallDistEntityQuerier(inst IService)
}
