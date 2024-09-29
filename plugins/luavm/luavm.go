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

package luavm

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/utils/option"
	"github.com/yuin/gopher-lua"
)

// ILuaVM lua虚拟机接口
type ILuaVM interface {
	GetLState() *lua.LState
}

func newLuaVM(setting ...option.Setting[LuaVMOptions]) ILuaVM {
	return &_LuaVM{
		options: option.Make(With.Default(), setting...),
	}
}

type _LuaVM struct {
	rtCtx   runtime.Context
	options LuaVMOptions
	ls      *lua.LState
}

// InitRP 初始化运行时插件
func (l *_LuaVM) InitRP(rtCtx runtime.Context) {
	l.rtCtx = rtCtx

	l.ls = lua.NewState(lua.Options{
		SkipOpenLibs:        !l.options.BuiltInLibraries,
		IncludeGoStackTrace: l.options.ShowGoStackTrace,
	})
}

// ShutRP 关闭运行时插件
func (l *_LuaVM) ShutRP(rtCtx runtime.Context) {
	if l.ls != nil {
		l.ls.Close()
	}
}

func (l *_LuaVM) GetLState() *lua.LState {
	return l.ls
}
