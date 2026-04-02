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

// Package framework provides the application bootstrap and assembly layer for
// Golaxy distributed services.
//
// 它在 git.golaxy.org/core 之上补齐了服务端开发常用的组合能力，包括：
//   - App 启动入口，以及基于 Cobra/Viper 的统一配置装载
//   - Service 与 Runtime 的默认 add-in 装配
//   - Service、Runtime、Entity、Component 之间的强类型桥接
//   - 生命周期接口、异步调用辅助，以及实体原型/实体构建器
//
// 典型使用方式是先创建 App，注册一个或多个服务装配器，再在服务生命周期
// 回调中创建 runtime、声明实体原型并生成实体实例。
package framework
