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

// InstallRuntimeLogger 为运行时提供自定义日志 add-in 安装钩子。
// 仅当 Birth 阶段尚未安装同名 add-in 时调用；实现必须在返回前完成安装。
type InstallRuntimeLogger interface {
	InstallLogger(rt IRuntime)
}

// InstallRuntimeRPCStack 为运行时提供自定义 RPC 调用栈 add-in 安装钩子。
// 仅当 Birth 阶段尚未安装同名 add-in 时调用；实现必须在返回前完成安装。
type InstallRuntimeRPCStack interface {
	InstallRPCStack(rt IRuntime)
}

// InstallRuntimeDistEntityRegistry 为运行时提供自定义分布式实体注册 add-in 安装钩子。
// 仅当 Birth 阶段尚未安装同名 add-in 时调用；实现必须在返回前完成安装。
type InstallRuntimeDistEntityRegistry interface {
	InstallDistEntityRegistry(rt IRuntime)
}
