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

// Name 返回不含 ETCD 键前缀的逻辑锁名称。
func (m *_EtcdSyncMutex) Name() string {
	return strings.TrimPrefix(m.name, m.dsync.options.KeyPrefix)
}

// UID 返回当前 ETCD session 的租约 ID；尚未创建 session 时返回空字符串。
func (m *_EtcdSyncMutex) UID() string {
	if m.session == nil {
		return ""
	}
	return strconv.Itoa(int(m.session.Lease()))
}

// Until 返回零值时间；ETCD 实现不提供本地租约截止时间。
func (m *_EtcdSyncMutex) Until() time.Time {
	log.L(m.dsync.svcCtx).Error("etcd mutex does not support retrieving the lock's expiration time")
	return time.Time{}
}

// TryLock 创建租约并尝试一次非阻塞加锁；当前句柄已在加锁时返回 ErrAlreadyAcquired。
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

// Lock 创建租约并等待获取锁，等待时间最多为配置的 Expiry。
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

// Unlock 释放 ETCD 锁并关闭租约 session；当前句柄未持锁时返回 ErrNotAcquired。
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

// Extend 始终返回不支持错误；ETCD session 会自行保持租约，无需手动续期。
func (m *_EtcdSyncMutex) Extend(ctx context.Context) error {
	log.L(m.dsync.svcCtx).Error("etcd mutex does not support extending the lock's expiration time")
	return errors.New("dsync: not supported")
}
