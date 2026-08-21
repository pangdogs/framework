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

// InstallServiceLogger 为服务提供自定义日志 add-in 安装钩子。
// 仅当 Birth 阶段尚未安装同名 add-in 时调用；实现必须在返回前完成安装。
// 若日志需供 OnTerminated 使用，安装的实例还应实现 service.RetainedAddIn。
type InstallServiceLogger interface {
	InstallLogger(svc IService)
}

// InstallServiceConfig 为服务提供自定义配置 add-in 安装钩子。
// 仅当 Birth 阶段尚未安装同名 add-in 时调用；实现必须在返回前完成安装。
// 若配置需供 OnTerminated 使用，安装的实例还应实现 service.RetainedAddIn。
type InstallServiceConfig interface {
	InstallConfig(svc IService)
}

// InstallServiceBroker 为服务提供自定义消息代理 add-in 安装钩子。
// 仅当 Birth 阶段尚未安装同名 add-in 时调用；实现必须在返回前完成安装。
type InstallServiceBroker interface {
	InstallBroker(svc IService)
}

// InstallServiceRegistry 为服务提供自定义服务发现 add-in 安装钩子。
// 仅当 Birth 阶段尚未安装同名 add-in 时调用；实现必须在返回前完成安装。
type InstallServiceRegistry interface {
	InstallRegistry(svc IService)
}

// InstallServiceDistSync 为服务提供自定义分布式同步 add-in 安装钩子。
// 仅当 Birth 阶段尚未安装同名 add-in 时调用；实现必须在返回前完成安装。
type InstallServiceDistSync interface {
	InstallDistSync(svc IService)
}

// InstallServiceDistService 为服务提供自定义分布式服务 add-in 安装钩子。
// 仅当 Birth 阶段尚未安装同名 add-in 时调用；实现必须在返回前完成安装。
type InstallServiceDistService interface {
	InstallDistService(svc IService)
}

// InstallServiceRPC 为服务提供自定义 RPC add-in 安装钩子。
// 仅当 Birth 阶段尚未安装同名 add-in 时调用；实现必须在返回前完成安装。
type InstallServiceRPC interface {
	InstallRPC(svc IService)
}

// InstallServiceDistEntityQuerier 为服务提供自定义分布式实体查询 add-in 安装钩子。
// 仅当 Birth 阶段尚未安装同名 add-in 时调用；实现必须在返回前完成安装。
type InstallServiceDistEntityQuerier interface {
	InstallDistEntityQuerier(svc IService)
}
