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

package etcd_dsync

import (
	"context"
	"errors"
	"fmt"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/dsync"
	"git.golaxy.org/framework/plugins/log"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcd_concurrency "go.etcd.io/etcd/client/v3/concurrency"
	"math"
	"strconv"
	"strings"
	"time"
)

type _DistMutexSettings struct {
	dsync *_DistSync
	name  string
}

// With applies additional settings to the distributed mutex.
func (s *_DistMutexSettings) With(settings ...option.Setting[dsync.DistMutexOptions]) dsync.IDistMutex {
	return s.dsync.NewMutex(s.name, settings...)
}

func (s *_DistSync) newMutex(name string, options dsync.DistMutexOptions) *_DistMutex {
	if s.options.KeyPrefix != "" {
		name = s.options.KeyPrefix + name
	}

	log.Debugf(s.svcCtx, "new dist mutex %q", name)

	return &_DistMutex{
		dsync:         s,
		name:          name,
		expiry:        options.Expiry,
		driftFactor:   options.DriftFactor,
		timeoutFactor: options.TimeoutFactor,
	}
}

type _DistMutex struct {
	dsync         *_DistSync
	name          string
	expiry        time.Duration
	driftFactor   float64
	timeoutFactor float64
	session       *etcd_concurrency.Session
	mutex         *etcd_concurrency.Mutex
	until         time.Time
}

// Name returns mutex name.
func (m *_DistMutex) Name() string {
	return strings.TrimPrefix(m.name, m.dsync.options.KeyPrefix)
}

// Value returns the current random value. The value will be empty until a lock is acquired (or Value option is used).
func (m *_DistMutex) Value() string {
	if m.session == nil {
		return ""
	}
	return strconv.Itoa(int(m.session.Lease()))
}

// Until returns the time of validity of acquired lock. The value will be zero value until a lock is acquired.
func (m *_DistMutex) Until() time.Time {
	return m.until
}

// Lock locks m. In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *_DistMutex) Lock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	expirySec := m.expiry.Seconds()
	if expirySec <= 0 {
		expirySec = 1
	}

	session, err := etcd_concurrency.NewSession(m.dsync.client, etcd_concurrency.WithTTL(int(math.Ceil(expirySec))))
	if err != nil {
		return fmt.Errorf("dsync: %w", err)
	}

	mutex := etcd_concurrency.NewMutex(session, m.name)

	start := time.Now()
	ctx, _ = context.WithTimeout(ctx, time.Duration((expirySec*m.timeoutFactor)*float64(time.Second)))

	if err = mutex.Lock(ctx); err != nil {
		session.Close()
		return fmt.Errorf("dsync: %w", err)
	}

	if _, err = m.dsync.client.KeepAlive(ctx, session.Lease()); err != nil {
		mutex.Unlock(context.Background())
		session.Close()
		return fmt.Errorf("dsync: %w", err)
	}

	m.clean()

	m.session = session
	m.mutex = mutex

	now := time.Now()
	m.until = now.Add(m.expiry - now.Sub(start) - time.Duration(int64(expirySec*m.driftFactor)))

	log.Debugf(m.dsync.svcCtx, "dist mutex %q is locked", m.name)

	return nil
}

// Unlock unlocks m and returns the status of unlock.
func (m *_DistMutex) Unlock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if m.mutex == nil {
		return dsync.ErrNotAcquired
	}

	defer m.clean()

	if err := m.mutex.Unlock(ctx); err != nil {
		if errors.Is(err, rpctypes.ErrKeyNotFound) {
			return dsync.ErrNotAcquired
		}
		return fmt.Errorf("dsync: %w", err)
	}

	log.Debugf(m.dsync.svcCtx, "dist mutex %q is unlocked", m.name)

	return nil
}

// Extend resets the mutex's expiry and returns the status of expiry extension.
func (m *_DistMutex) Extend(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if m.session == nil {
		return dsync.ErrNotAcquired
	}

	if _, err := m.dsync.client.KeepAlive(ctx, m.session.Lease()); err != nil {
		if errors.Is(err, rpctypes.ErrLeaseNotFound) {
			return dsync.ErrNotAcquired
		}
		return fmt.Errorf("dsync: %w", err)
	}

	log.Debugf(m.dsync.svcCtx, "dist mutex %q is extended", m.name)

	return nil
}

// Valid returns true if the lock acquired through m is still valid. It may
// also return true erroneously if quorum is achieved during the call and at
// least one node then takes long enough to respond for the lock to expire.
func (m *_DistMutex) Valid(ctx context.Context) (bool, error) {
	if m.session == nil {
		return false, nil
	}

	select {
	case <-m.session.Done():
		return false, nil
	default:
		return true, nil
	}
}

func (m *_DistMutex) clean() {
	if m.session != nil {
		m.session.Close()
	}
	m.session = nil
	m.mutex = nil
	m.until = time.Time{}
}
