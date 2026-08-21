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

// Map 建立实体与会话的一对一映射；任一端已有其他映射时，旧映射会先被移除。
func (r *_Router) Map(entityID, sessionID uid.ID) (IMapping, error) {
	select {
	case <-r.scope.Context().Done():
		return nil, errors.New("router: router is terminating")
	default:
	}

	if !r.barrier.Join(1) {
		return nil, errors.New("router: router is terminating")
	}

	entity, ok := r.svcCtx.EntityManager().GetEntity(entityID)
	if !ok {
		r.barrier.Done()
		return nil, ErrEntityNotFound
	}

	session, ok := r.gate.Get(sessionID)
	if !ok {
		r.barrier.Done()
		return nil, ErrSessionNotFound
	}

	removed, _ := async.NewSignal()
	unmapped, _ := async.NewSignal()
	mapping := &_Mapping{
		router:     r,
		clientAddr: gate.ClientDetails.DomainUnicast.Join(entity.ID().String()),
		entity:     entity,
		session:    session,
		removed:    removed,
		unmapped:   unmapped,
	}

	r.mappingMu.Lock()

	currByEntity := r.mappings[entityID]
	currBySession := r.mappings[sessionID]

	if currByEntity != nil && currByEntity == currBySession {
		r.mappingMu.Unlock()
		r.barrier.Done()
		return currByEntity, nil
	}

	if currByEntity != nil {
		if r.removeMapping(currByEntity) {
			currByEntity.removed.Complete()
		}
	}

	if currBySession != nil {
		if r.removeMapping(currBySession) {
			currBySession.removed.Complete()
		}
	}

	r.mappings[entity.ID()] = mapping
	r.mappings[session.ID()] = mapping

	r.mappingMu.Unlock()

	go mapping.waitForUnmap()

	log.L(r.svcCtx).Info("add mapping",
		zap.String("entity_id", entity.ID().String()),
		zap.String("session_id", session.ID().String()))

	return mapping, nil
}

// Lookup 按实体 ID 或会话 ID 查询当前映射。
func (r *_Router) Lookup(id uid.ID) (IMapping, bool) {
	mapping, ok := r.getMappingLocked(id)
	if !ok {
		return nil, false
	}
	return mapping, true
}

func (r *_Router) getMappingLocked(id uid.ID) (*_Mapping, bool) {
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
	if curr := r.mappings[m.entity.ID()]; curr == m {
		delete(r.mappings, m.entity.ID())
		removed = true
	}
	if curr := r.mappings[m.session.ID()]; curr == m {
		delete(r.mappings, m.session.ID())
		removed = true
	}
	return removed
}
