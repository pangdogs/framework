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
	"sync"

	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"go.uber.org/zap"
)

// IMapping 路由映射接口
type IMapping interface {
	// ClientAddr 获取客户端地址
	ClientAddr() string
	// Entity 获取实体
	Entity() ec.ConcurrentEntity
	// Session 获取会话
	Session() gate.ISession
	// Unmap 取消映射
	Unmap()
	// Unmapped 已取消映射
	Unmapped() async.Future
}

type _Mapping struct {
	router     *_Router
	clientAddr string
	entity     ec.ConcurrentEntity
	session    gate.ISession
	unmapOnce  sync.Once
	removed    async.FutureVoid
	unmapped   async.FutureVoid
}

// ClientAddr 获取客户端地址
func (m *_Mapping) ClientAddr() string {
	return m.clientAddr
}

// Entity 获取实体
func (m *_Mapping) Entity() ec.ConcurrentEntity {
	return m.entity
}

// Session 获取会话
func (m *_Mapping) Session() gate.ISession {
	return m.session
}

// Unmap 取消映射
func (m *_Mapping) Unmap() {
	m.unmapOnce.Do(func() {
		if m.router.removeMappingLocked(m) {
			async.ReturnVoid(m.removed)
		}
	})
}

// Unmapped 已取消映射
func (m *_Mapping) Unmapped() async.Future {
	return m.unmapped.Out()
}

func (m *_Mapping) waitForUnmap() {
	var reason string

	defer func() {
		log.L(m.router.svcCtx).Info("mapping unmapped",
			zap.String("entity_id", m.entity.Id().String()),
			zap.String("session_id", m.session.Id().String()),
			zap.String("reason", reason))

		async.ReturnVoid(m.unmapped)
		m.router.barrier.Done()
	}()

	select {
	case <-m.router.ctx.Done():
		reason = "router_terminating"
	case <-m.entity.Terminated().Done():
		reason = "entity_destroyed"
	case <-m.session.Closed().Done():
		reason = "session_closed"
	case <-m.removed:
		reason = "mapping_removed"
		return
	}

	m.Unmap()
}
