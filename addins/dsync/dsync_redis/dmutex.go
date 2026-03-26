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

package dsync_redis

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	"git.golaxy.org/framework/addins/dsync"
	"git.golaxy.org/framework/addins/log"
	"github.com/go-redsync/redsync/v4"
	"go.uber.org/zap"
)

func (s *_RedisSync) newMutex(name string, options dsync.DistMutexOptions) *_RedisSyncMutex {
	if s.options.KeyPrefix != "" {
		name = s.options.KeyPrefix + name
	}

	mutex := s.redSync.NewMutex(name,
		redsync.WithExpiry(options.Expiry),
		redsync.WithTries(options.Tries),
		redsync.WithRetryDelayFunc(redsync.DelayFunc(options.RetryDelayFunc)),
		redsync.WithDriftFactor(options.DriftFactor),
		redsync.WithTimeoutFactor(options.TimeoutFactor),
		redsync.WithGenValueFunc(options.GenUIDFunc),
		redsync.WithValue(options.UID),
	)

	log.L(s.svcCtx).Debug("redis mutex created", zap.String("name", mutex.Name()), zap.String("uid", mutex.Value()))

	return &_RedisSyncMutex{
		dsync: s,
		Mutex: mutex,
	}
}

type _RedisSyncMutex struct {
	dsync *_RedisSync
	*redsync.Mutex
	locked atomic.Bool
}

// Name 名称
func (m *_RedisSyncMutex) Name() string {
	return strings.TrimPrefix(m.Mutex.Name(), m.dsync.options.KeyPrefix)
}

// UID 唯一ID
func (m *_RedisSyncMutex) UID() string {
	return m.Value()
}

// TryLock 尝试加锁，支持错误重试
func (m *_RedisSyncMutex) TryLock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if !m.locked.CompareAndSwap(false, true) {
		log.L(m.dsync.svcCtx).Debug("redis mutex already locked",
			zap.String("name", m.Mutex.Name()),
			zap.String("uid", m.Mutex.Value()))

		return dsync.ErrAlreadyAcquired
	}

	if err := m.TryLockContext(ctx); err != nil {
		m.locked.Store(false)

		log.L(m.dsync.svcCtx).Error("redis mutex try lock failed",
			zap.String("name", m.Mutex.Name()),
			zap.String("uid", m.Mutex.Value()))

		return fmt.Errorf("dsync: %w", err)
	}

	log.L(m.dsync.svcCtx).Debug("redis mutex lock acquired",
		zap.String("name", m.Mutex.Name()),
		zap.String("uid", m.Mutex.Value()))

	return nil
}

// Lock 加锁，支持错误重试
func (m *_RedisSyncMutex) Lock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if !m.locked.CompareAndSwap(false, true) {
		log.L(m.dsync.svcCtx).Debug("redis mutex already locked",
			zap.String("name", m.Mutex.Name()),
			zap.String("uid", m.Mutex.Value()))

		return dsync.ErrAlreadyAcquired
	}

	if err := m.LockContext(ctx); err != nil {
		m.locked.Store(false)

		log.L(m.dsync.svcCtx).Error("redis mutex lock failed",
			zap.String("name", m.Mutex.Name()),
			zap.String("uid", m.Mutex.Value()))

		return fmt.Errorf("dsync: %w", err)
	}

	log.L(m.dsync.svcCtx).Debug("redis mutex lock acquired",
		zap.String("name", m.Mutex.Name()),
		zap.String("uid", m.Mutex.Value()))

	return nil
}

// Unlock 解锁
func (m *_RedisSyncMutex) Unlock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if !m.locked.CompareAndSwap(true, false) {
		log.L(m.dsync.svcCtx).Debug("redis mutex lock not acquired",
			zap.String("name", m.Mutex.Name()),
			zap.String("uid", m.Mutex.Value()))

		return dsync.ErrNotAcquired
	}

	ok, err := m.UnlockContext(ctx)
	if err != nil {
		log.L(m.dsync.svcCtx).Error("redis mutex unlock failed",
			zap.String("name", m.Mutex.Name()),
			zap.String("uid", m.Mutex.Value()))

		return fmt.Errorf("dsync: %w", err)
	}
	if !ok {
		return dsync.ErrNotAcquired
	}

	log.L(m.dsync.svcCtx).Debug("redis mutex lock released",
		zap.String("name", m.Mutex.Name()),
		zap.String("uid", m.Mutex.Value()))

	return nil
}

// Extend 延长锁的过期时间
func (m *_RedisSyncMutex) Extend(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	ok, err := m.ExtendContext(ctx)
	if err != nil {
		log.L(m.dsync.svcCtx).Error("redis mutex lock extend failed",
			zap.String("name", m.Mutex.Name()),
			zap.String("uid", m.Mutex.Value()))

		return fmt.Errorf("dsync: %w", err)
	}
	if !ok {
		return dsync.ErrNotAcquired
	}

	log.L(m.dsync.svcCtx).Debug("redis mutex lock extended",
		zap.String("name", m.Mutex.Name()),
		zap.String("uid", m.Mutex.Value()))

	return nil
}
