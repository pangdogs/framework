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

import "git.golaxy.org/core/ec"

// LifecycleServiceBirth 在服务上下文及基础配置创建后、framework add-in 装配前接收回调。
type LifecycleServiceBirth interface {
	OnBirth(svc IService)
}

// LifecycleServiceBuilt 在 framework add-in 装配完成后接收回调。
// 这是服务进入 Starting 并冻结 add-in 管理器前补充服务 add-in 的最后阶段。
type LifecycleServiceBuilt interface {
	OnBuilt(svc IService)
}

// LifecycleServiceStarting 在服务开始运行时接收回调。
// 服务 add-in 已在回调前冻结并激活，之后不能再安装或卸载。
type LifecycleServiceStarting interface {
	OnStarting(svc IService)
}

// LifecycleServiceStarted 在服务完成启动并加入分布式网络后接收回调。
type LifecycleServiceStarted interface {
	OnStarted(svc IService)
}

// LifecycleServiceHeartbeat 在服务运行期间接收每秒一次的心跳回调。
type LifecycleServiceHeartbeat interface {
	OnHeartbeat(svc IService)
}

// LifecycleServiceTerminating 在服务开始停止、等待子任务退出前接收回调。
type LifecycleServiceTerminating interface {
	OnTerminating(svc IService)
}

// LifecycleServiceTerminated 在服务等待组清空且 add-in 停用后接收回调。
type LifecycleServiceTerminated interface {
	OnTerminated(svc IService)
}

// LifecycleServiceEntityPTDeclared 在服务实体库声明实体原型后接收回调。
type LifecycleServiceEntityPTDeclared interface {
	OnEntityPTDeclared(svc IService, entityPT ec.EntityPT)
}

// LifecycleServiceComponentPTDeclared 在服务组件库声明组件原型后接收回调。
type LifecycleServiceComponentPTDeclared interface {
	OnComponentPTDeclared(svc IService, componentPT ec.ComponentPT)
}

// LifecycleServiceEntityRegistered 在全局实体注册到服务实体库后接收回调。
type LifecycleServiceEntityRegistered interface {
	OnEntityRegistered(svc IService, entity ec.ConcurrentEntity)
}

// LifecycleServiceEntityDeregistered 在全局实体从服务实体库注销后接收回调。
type LifecycleServiceEntityDeregistered interface {
	OnEntityDeregistered(svc IService, entity ec.ConcurrentEntity)
}
