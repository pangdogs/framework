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
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/reinterpret"
)

// EntityBehavior 实体行为，在需要扩展实体能力时，匿名嵌入至实体结构体中
type EntityBehavior struct {
	ec.EntityBehavior
}

// GetRuntime 获取运行时
func (e *EntityBehavior) GetRuntime() IRuntimeInstance {
	return reinterpret.Cast[IRuntimeInstance](runtime.Current(e))
}

// GetService 获取服务
func (e *EntityBehavior) GetService() IServiceInstance {
	return reinterpret.Cast[IServiceInstance](service.Current(e))
}

// GetPluginBundle 获取插件包
func (e *EntityBehavior) GetPluginBundle() extension.PluginBundle {
	return runtime.Current(e).GetPluginBundle()
}

// IsAlive 是否活跃
func (e *EntityBehavior) IsAlive() bool {
	return e.GetState() <= ec.EntityState_Alive
}
