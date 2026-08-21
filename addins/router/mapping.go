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

// IMapping 表示一个实体、网关会话和客户端地址之间的一对一映射。
type IMapping interface {
	// ClientAddr 返回该会话的客户端单播地址。
	ClientAddr() string
	// Entity 返回被映射的并发实体。
	Entity() ec.ConcurrentEntity
	// Session 返回被映射的网关会话。
	Session() gate.ISession
	// Unmap 请求取消映射；重复调用无副作用。
	Unmap()
	// Unmapped 返回映射完全取消后完成的 Signal。
	Unmapped() async.Signal
}

type _Mapping struct {
	router     *_Router
	clientAddr string
	entity     ec.ConcurrentEntity
	session    gate.ISession
	unmapOnce  sync.Once
	removed    async.Completer
	unmapped   async.Completer
}

// ClientAddr 返回该会话的客户端单播地址。
func (m *_Mapping) ClientAddr() string {
	return m.clientAddr
}

// Entity 返回被映射的并发实体。
func (m *_Mapping) Entity() ec.ConcurrentEntity {
	return m.entity
}

// Session 返回被映射的网关会话。
func (m *_Mapping) Session() gate.ISession {
	return m.session
}

// Unmap 请求取消映射；重复调用无副作用。
func (m *_Mapping) Unmap() {
	m.unmapOnce.Do(func() {
		if m.router.removeMappingLocked(m) {
			m.removed.Complete()
		}
	})
}

// Unmapped 返回映射完全取消后完成的 Signal。
func (m *_Mapping) Unmapped() async.Signal {
	return m.unmapped.Signal()
}

func (m *_Mapping) waitForUnmap() {
	defer m.router.barrier.Done()

	var reason string

	select {
	case <-m.router.scope.Context().Done():
		m.Unmap()
		reason = "router_terminating"
	case <-m.entity.Terminated().Done():
		m.Unmap()
		reason = "entity_destroyed"
	case <-m.session.Closed().Done():
		m.Unmap()
		reason = "session_closed"
	case <-m.removed.Signal().Done():
		reason = "mapping_removed"
	}

	log.L(m.router.svcCtx).Info("mapping unmapped",
		zap.String("entity_id", m.entity.ID().String()),
		zap.String("session_id", m.session.ID().String()),
		zap.String("reason", reason))

	m.unmapped.Complete()
}
