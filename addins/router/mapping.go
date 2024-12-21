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

package router

import (
	"context"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/framework/addins/gate"
)

// IMapping 映射
type IMapping interface {
	context.Context
	// GetEntity 获取实体
	GetEntity() ec.ConcurrentEntity
	// GetSession 获取会话
	GetSession() gate.ISession
	// GetCliAddr 获取客户端地址
	GetCliAddr() string
}

type _Mapping struct {
	context.Context
	terminate context.CancelFunc
	router    *_Router
	entity    ec.ConcurrentEntity
	session   gate.ISession
	cliAddr   string
}

// GetEntity 获取实体
func (m *_Mapping) GetEntity() ec.ConcurrentEntity {
	return m.entity
}

// GetSession 获取会话
func (m *_Mapping) GetSession() gate.ISession {
	return m.session
}

// GetCliAddr 获取客户端地址
func (m *_Mapping) GetCliAddr() string {
	return m.cliAddr
}

func (m *_Mapping) mainLoop() {
	select {
	case <-m.Done():
		return
	case <-m.entity.Done():
		m.router.CleanEntity(m.entity.GetId())
	case <-m.session.Done():
		m.router.CleanSession(m.session.GetId())
	}
}
