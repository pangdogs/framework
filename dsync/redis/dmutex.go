package redis

import (
	"context"
	"github.com/go-redsync/redsync/v4"
	"kit.golaxy.org/plugins/dsync"
	"kit.golaxy.org/plugins/logger"
	"strings"
)

func newRedisDMutex(rs *_RedisDsync, name string, options dsync.DMutexOptions) dsync.DMutex {
	if rs.options.KeyPrefix != "" {
		name = rs.options.KeyPrefix + name
	}

	rMutex := rs.NewMutex(name,
		redsync.WithExpiry(options.Expiry),
		redsync.WithTries(options.Tries),
		redsync.WithRetryDelayFunc(options.DelayFunc),
		redsync.WithDriftFactor(options.DriftFactor),
		redsync.WithTimeoutFactor(options.TimeoutFactor),
		redsync.WithGenValueFunc(options.GenValueFunc),
		redsync.WithValue(options.Value),
	)

	logger.Debugf(rs.ctx, "new dmutex %q", name)

	return &_RedisDMutex{
		rs:    rs,
		Mutex: rMutex,
	}
}

type _RedisDMutex struct {
	rs *_RedisDsync
	*redsync.Mutex
}

// Name returns mutex name.
func (m *_RedisDMutex) Name() string {
	return strings.TrimPrefix(m.Mutex.Name(), m.rs.options.KeyPrefix)
}

// Lock locks m. In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *_RedisDMutex) Lock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	err := m.LockContext(ctx)
	if err != nil {
		return err
	}

	logger.Debugf(m.rs.ctx, "dmutex %q is locked", m.Mutex.Name())

	return nil
}

// Unlock unlocks m and returns the status of unlock.
func (m *_RedisDMutex) Unlock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	ok, err := m.UnlockContext(ctx)
	if err != nil {
		return err
	}

	if !ok {
		return dsync.ErrNotAcquired
	}

	logger.Debugf(m.rs.ctx, "dmutex %q is unlocked", m.Mutex.Name())

	return nil
}

// Extend resets the mutex's expiry and returns the status of expiry extension.
func (m *_RedisDMutex) Extend(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	ok, err := m.ExtendContext(ctx)
	if err != nil {
		return err
	}

	if !ok {
		return dsync.ErrNotAcquired
	}

	logger.Debugf(m.rs.ctx, "dmutex %q is extended", m.Mutex.Name())

	return nil
}

// Valid returns true if the lock acquired through m is still valid. It may
// also return true erroneously if quorum is achieved during the call and at
// least one node then takes long enough to respond for the lock to expire.
func (m *_RedisDMutex) Valid(ctx context.Context) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	return m.ValidContext(ctx)
}
