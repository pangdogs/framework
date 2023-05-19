package redis

import (
	"github.com/go-redsync/redsync/v4"
	"golang.org/x/net/context"
	"kit.golaxy.org/plugins/dsync"
	"strings"
)

func newRedisDMutex(rs *_RedisDsync, name string, options dsync.Options) dsync.DMutex {
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
	return m.LockContext(ctx)
}

// Unlock unlocks m and returns the status of unlock.
func (m *_RedisDMutex) Unlock(ctx context.Context) error {
	ok, err := m.UnlockContext(ctx)
	if err != nil {
		return err
	}

	if !ok {
		return dsync.ErrNotObtained
	}

	return nil
}

// Extend resets the mutex's expiry and returns the status of expiry extension.
func (m *_RedisDMutex) Extend(ctx context.Context) error {
	ok, err := m.ExtendContext(ctx)
	if err != nil {
		return err
	}

	if !ok {
		return dsync.ErrNotObtained
	}

	return nil
}

// Valid returns true if the lock acquired through m is still valid. It may
// also return true erroneously if quorum is achieved during the call and at
// least one node then takes long enough to respond for the lock to expire.
func (m *_RedisDMutex) Valid(ctx context.Context) (bool, error) {
	return m.ValidContext(ctx)
}
