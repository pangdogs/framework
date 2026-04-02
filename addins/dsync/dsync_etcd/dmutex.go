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

package dsync_etcd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"git.golaxy.org/framework/addins/dsync"
	"git.golaxy.org/framework/addins/log"
	etcd_concurrency "go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"
)

func (s *_EtcdSync) newMutex(name string, options dsync.DistMutexOptions) *_EtcdSyncMutex {
	if s.options.KeyPrefix != "" {
		name = s.options.KeyPrefix + name
	}

	if options.UID != "" {
		log.L(s.svcCtx).Warn("etcd mutex does not support specifying a UID")
	}

	log.L(s.svcCtx).Debug("etcd mutex created", zap.String("name", name))

	return &_EtcdSyncMutex{
		dsync:  s,
		name:   name,
		expiry: options.Expiry,
	}
}

type _EtcdSyncMutex struct {
	dsync   *_EtcdSync
	name    string
	expiry  time.Duration
	session *etcd_concurrency.Session
	mutex   *etcd_concurrency.Mutex
	locked  atomic.Bool
}

// Name 名称
func (m *_EtcdSyncMutex) Name() string {
	return strings.TrimPrefix(m.name, m.dsync.options.KeyPrefix)
}

// UID 唯一ID
func (m *_EtcdSyncMutex) UID() string {
	if m.session == nil {
		return ""
	}
	return strconv.Itoa(int(m.session.Lease()))
}

// Until 返回锁的有效期结束时间
func (m *_EtcdSyncMutex) Until() time.Time {
	log.L(m.dsync.svcCtx).Error("etcd mutex does not support retrieving the lock's expiration time")
	return time.Time{}
}

// TryLock 尝试加锁，支持错误重试
func (m *_EtcdSyncMutex) TryLock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if !m.locked.CompareAndSwap(false, true) {
		log.L(m.dsync.svcCtx).Debug("etcd mutex already locked", zap.String("name", m.name))
		return dsync.ErrAlreadyAcquired
	}

	session, err := etcd_concurrency.NewSession(m.dsync.client, etcd_concurrency.WithTTL(int(math.Ceil(m.expiry.Seconds()))))
	if err != nil {
		m.locked.Store(false)

		log.L(m.dsync.svcCtx).Error("etcd mutex create session failed", zap.String("name", m.name), zap.Error(err))
		return fmt.Errorf("dsync: %w", err)
	}

	mutex := etcd_concurrency.NewMutex(session, m.name)

	if err = mutex.TryLock(ctx); err != nil {
		session.Close()
		m.locked.Store(false)

		log.L(m.dsync.svcCtx).Error("etcd mutex try lock failed", zap.String("name", m.name), zap.Int64("lease_id", int64(session.Lease())), zap.Error(err))
		return fmt.Errorf("dsync: %w", err)
	}

	m.session = session
	m.mutex = mutex

	log.L(m.dsync.svcCtx).Debug("etcd mutex lock acquired",
		zap.String("name", m.name),
		zap.Int64("lease_id", int64(session.Lease())))

	return nil
}

// Lock 加锁，支持错误重试
func (m *_EtcdSyncMutex) Lock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if !m.locked.CompareAndSwap(false, true) {
		log.L(m.dsync.svcCtx).Debug("etcd mutex already locked", zap.String("name", m.name))
		return dsync.ErrAlreadyAcquired
	}

	session, err := etcd_concurrency.NewSession(m.dsync.client, etcd_concurrency.WithTTL(int(math.Ceil(m.expiry.Seconds()))))
	if err != nil {
		m.locked.Store(false)

		log.L(m.dsync.svcCtx).Error("etcd mutex create session failed", zap.String("name", m.name), zap.Error(err))
		return fmt.Errorf("dsync: %w", err)
	}

	lockCtx, cancel := context.WithTimeout(ctx, m.expiry)
	defer cancel()

	mutex := etcd_concurrency.NewMutex(session, m.name)

	if err = mutex.Lock(lockCtx); err != nil {
		session.Close()
		m.locked.Store(false)

		log.L(m.dsync.svcCtx).Error("etcd mutex lock failed", zap.String("name", m.name), zap.Int64("lease_id", int64(session.Lease())), zap.Error(err))
		return fmt.Errorf("dsync: %w", err)
	}

	m.session = session
	m.mutex = mutex

	log.L(m.dsync.svcCtx).Debug("etcd mutex lock acquired",
		zap.String("name", m.name),
		zap.Int64("lease_id", int64(session.Lease())))

	return nil
}

// Unlock 解锁
func (m *_EtcdSyncMutex) Unlock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if !m.locked.CompareAndSwap(true, false) {
		log.L(m.dsync.svcCtx).Debug("etcd mutex lock not acquired", zap.String("name", m.name))
		return dsync.ErrNotAcquired
	}

	defer m.session.Close()

	if err := m.mutex.Unlock(ctx); err != nil {
		log.L(m.dsync.svcCtx).Error("etcd mutex unlock failed", zap.String("name", m.name), zap.Int64("lease_id", int64(m.session.Lease())), zap.Error(err))
		return fmt.Errorf("dsync: %w", err)
	}

	log.L(m.dsync.svcCtx).Debug("etcd mutex lock released",
		zap.String("name", m.name),
		zap.Int64("lease_id", int64(m.session.Lease())))

	return nil
}

// Extend 延长锁的过期时间
func (m *_EtcdSyncMutex) Extend(ctx context.Context) error {
	log.L(m.dsync.svcCtx).Error("etcd mutex does not support extending the lock's expiration time")
	return errors.New("dsync: not supported")
}
