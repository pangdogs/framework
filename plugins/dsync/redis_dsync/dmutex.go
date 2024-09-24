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

package redis_dsync

import (
	"context"
	"fmt"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/dsync"
	"git.golaxy.org/framework/plugins/log"
	"github.com/go-redsync/redsync/v4"
	"strings"
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

	mutex := s.redSync.NewMutex(name,
		redsync.WithExpiry(options.Expiry),
		redsync.WithTries(options.Tries),
		redsync.WithRetryDelayFunc(redsync.DelayFunc(options.DelayFunc)),
		redsync.WithDriftFactor(options.DriftFactor),
		redsync.WithTimeoutFactor(options.TimeoutFactor),
		redsync.WithGenValueFunc(options.GenValueFunc),
		redsync.WithValue(options.Value),
	)

	log.Debugf(s.svcCtx, "new dist mutex %q", name)

	return &_DistMutex{
		dsync: s,
		Mutex: mutex,
	}
}

type _DistMutex struct {
	dsync *_DistSync
	*redsync.Mutex
}

// Name returns mutex name.
func (m *_DistMutex) Name() string {
	return strings.TrimPrefix(m.Mutex.Name(), m.dsync.options.KeyPrefix)
}

// Lock locks m. In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *_DistMutex) Lock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := m.LockContext(ctx); err != nil {
		return fmt.Errorf("dsync: %w", err)
	}

	log.Debugf(m.dsync.svcCtx, "dist mutex %q is locked", m.Mutex.Name())

	return nil
}

// Unlock unlocks m and returns the status of unlock.
func (m *_DistMutex) Unlock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	ok, err := m.UnlockContext(ctx)
	if err != nil {
		return fmt.Errorf("dsync: %w", err)
	}

	if !ok {
		return dsync.ErrNotAcquired
	}

	log.Debugf(m.dsync.svcCtx, "dist mutex %q is unlocked", m.Mutex.Name())

	return nil
}

// Extend resets the mutex's expiry and returns the status of expiry extension.
func (m *_DistMutex) Extend(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	ok, err := m.ExtendContext(ctx)
	if err != nil {
		return fmt.Errorf("dsync: %w", err)
	}

	if !ok {
		return dsync.ErrNotAcquired
	}

	log.Debugf(m.dsync.svcCtx, "dist mutex %q is extended", m.Mutex.Name())

	return nil
}

// Valid returns true if the lock acquired through m is still valid. It may
// also return true erroneously if quorum is achieved during the call and at
// least one node then takes long enough to respond for the lock to expire.
func (m *_DistMutex) Valid(ctx context.Context) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	b, err := m.ValidContext(ctx)
	if err != nil {
		return b, fmt.Errorf("dsync: %w", err)
	}

	return b, nil
}
