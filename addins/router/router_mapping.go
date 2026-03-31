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
	"errors"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"go.uber.org/zap"
)

// Map 添加路由映射
func (r *_Router) Map(entityId, sessionId uid.Id) (IMapping, error) {
	select {
	case <-r.ctx.Done():
		return nil, errors.New("router: router is terminating")
	default:
	}

	if !r.barrier.Join(1) {
		return nil, errors.New("router: router is terminating")
	}

	entity, ok := r.svcCtx.EntityManager().GetEntity(entityId)
	if !ok {
		r.barrier.Done()
		return nil, ErrEntityNotFound
	}

	session, ok := r.gate.Get(sessionId)
	if !ok {
		r.barrier.Done()
		return nil, ErrSessionNotFound
	}

	mapping := &_Mapping{
		router:     r,
		clientAddr: gate.ClientDetails.DomainUnicast.Join(entity.Id().String()),
		entity:     entity,
		session:    session,
		replaced:   async.NewFutureVoid(),
		unmapped:   async.NewFutureVoid(),
	}

	r.mappingMu.Lock()

	currByEntity := r.mappings[entityId]
	currBySession := r.mappings[sessionId]

	if currByEntity != nil && currByEntity == currBySession {
		r.mappingMu.Unlock()
		r.barrier.Done()
		return currByEntity, nil
	}

	if currByEntity != nil {
		if r.removeMapping(currByEntity) {
			async.ReturnVoid(currByEntity.replaced)
		}
	}

	if currBySession != nil {
		if r.removeMapping(currBySession) {
			async.ReturnVoid(currBySession.replaced)
		}
	}

	r.mappings[entity.Id()] = mapping
	r.mappings[session.Id()] = mapping

	r.mappingMu.Unlock()

	go mapping.waitForExpire()

	log.L(r.svcCtx).Debug("add mapping",
		zap.String("entity_id", entity.Id().String()),
		zap.String("session_id", session.Id().String()))

	return mapping, nil
}

// Lookup 查询映射，可传实体id或会话id
func (r *_Router) Lookup(id uid.Id) (IMapping, bool) {
	mapping, ok := r.getMappingLocked(id)
	if !ok {
		return nil, false
	}
	return mapping, true
}

func (r *_Router) getMappingLocked(id uid.Id) (*_Mapping, bool) {
	r.mappingMu.RLock()
	mapping, ok := r.mappings[id]
	r.mappingMu.RUnlock()
	return mapping, ok
}

func (r *_Router) removeMappingLocked(m *_Mapping) bool {
	r.mappingMu.Lock()
	b := r.removeMapping(m)
	r.mappingMu.Unlock()
	return b
}

func (r *_Router) removeMapping(m *_Mapping) bool {
	removed := false
	if curr := r.mappings[m.entity.Id()]; curr == m {
		delete(r.mappings, m.entity.Id())
		removed = true
	}
	if curr := r.mappings[m.session.Id()]; curr == m {
		delete(r.mappings, m.session.Id())
		removed = true
	}
	return removed
}
