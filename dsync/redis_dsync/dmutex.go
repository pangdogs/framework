package redis_dsync

import (
	"context"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"kit.golaxy.org/plugins/dsync"
	"kit.golaxy.org/plugins/log"
	"strings"
)

func (s *_Dsync) newMutex(name string, options dsync.DMutexOptions) dsync.DMutex {
	if s.options.KeyPrefix != "" {
		name = s.options.KeyPrefix + name
	}

	mutex := s.redSync.NewMutex(name,
		redsync.WithExpiry(options.Expiry),
		redsync.WithTries(options.Tries),
		redsync.WithRetryDelayFunc(options.DelayFunc),
		redsync.WithDriftFactor(options.DriftFactor),
		redsync.WithTimeoutFactor(options.TimeoutFactor),
		redsync.WithGenValueFunc(options.GenValueFunc),
		redsync.WithValue(options.Value),
	)

	log.Debugf(s.ctx, "new dsync mutex %q", name)

	return &_DMutex{
		dsync: s,
		Mutex: mutex,
	}
}

type _DMutex struct {
	dsync *_Dsync
	*redsync.Mutex
}

// Name returns mutex name.
func (m *_DMutex) Name() string {
	return strings.TrimPrefix(m.Mutex.Name(), m.dsync.options.KeyPrefix)
}

// Lock locks m. In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *_DMutex) Lock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := m.LockContext(ctx); err != nil {
		return fmt.Errorf("%w: %w", dsync.ErrDsync, err)
	}

	log.Debugf(m.dsync.ctx, "dsync mutex %q is locked", m.Mutex.Name())

	return nil
}

// Unlock unlocks m and returns the status of unlock.
func (m *_DMutex) Unlock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	ok, err := m.UnlockContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", dsync.ErrDsync, err)
	}

	if !ok {
		return dsync.ErrNotAcquired
	}

	log.Debugf(m.dsync.ctx, "dsync mutex %q is unlocked", m.Mutex.Name())

	return nil
}

// Extend resets the mutex's expiry and returns the status of expiry extension.
func (m *_DMutex) Extend(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	ok, err := m.ExtendContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", dsync.ErrDsync, err)
	}

	if !ok {
		return dsync.ErrNotAcquired
	}

	log.Debugf(m.dsync.ctx, "dsync mutex %q is extended", m.Mutex.Name())

	return nil
}

// Valid returns true if the lock acquired through m is still valid. It may
// also return true erroneously if quorum is achieved during the call and at
// least one node then takes long enough to respond for the lock to expire.
func (m *_DMutex) Valid(ctx context.Context) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	b, err := m.ValidContext(ctx)
	if err != nil {
		return b, fmt.Errorf("%w: %w", dsync.ErrDsync, err)
	}

	return b, nil
}
