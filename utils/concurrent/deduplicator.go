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

package concurrent

import (
	"sync/atomic"
	"time"
)

// NewDeduplicator 创建去重器
func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		localSeq:     time.Now().UnixMicro(),
		remoteSeqMap: MakeLockedMap[string, *_RemoteSeq](0),
	}
}

type _RemoteSeq = Locked[int64]

// Deduplicator 去重器，用于保持幂等性
type Deduplicator struct {
	localSeq     int64
	remoteSeqMap LockedMap[string, *_RemoteSeq]
}

// Make 创建序号
func (d *Deduplicator) Make() int64 {
	return atomic.AddInt64(&d.localSeq, 1)
}

// Validate 验证序号
func (d *Deduplicator) Validate(remote string, seq int64) (passed bool) {
	remoteSeq, ok := d.remoteSeqMap.Get(remote)
	if !ok {
		var firstInsert bool

		d.remoteSeqMap.AutoLock(func(m *map[string]*_RemoteSeq) {
			remoteSeq, ok = (*m)[remote]
			if !ok {
				remoteSeq = NewLocked[int64](seq)
				(*m)[remote] = remoteSeq

				firstInsert = true
			}
		})

		if firstInsert {
			return true
		}
	}

	remoteSeq.AutoLock(func(remoteSeq *int64) {
		if seq <= *remoteSeq {
			return
		}
		*remoteSeq = seq
		passed = true
	})

	return
}

// Remove 删除对端
func (d *Deduplicator) Remove(remote string) {
	d.remoteSeqMap.Delete(remote)
}
