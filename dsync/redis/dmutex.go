package redis

import (
	"github.com/go-redsync/redsync/v4"
	"golang.org/x/net/context"
	"kit.golaxy.org/plugins/dsync"
)

func newRedisDMutex(s *_RedisDsync, name string, options dsync.Options) dsync.DMutex {
	if s.options.KeyPrefix != "" {
		name = s.options.KeyPrefix + name
	}

	rsMutex := s.rs.NewMutex(name,
		redsync.WithTries(options.Tries),
		redsync.WithExpiry(options.Expiry),
		redsync.WithRetryDelayFunc(redsync.DelayFunc(options.DelayFunc)),
		redsync.WithDriftFactor(options.DriftFactor),
		redsync.WithTimeoutFactor(options.TimeoutFactor),
		redsync.WithGenValueFunc(options.GenValueFunc),
		redsync.WithValue(options.Value),
	)

	return &_RedisDMutex{
		Mutex: rsMutex,
	}
}

type _RedisDMutex struct {
	*redsync.Mutex
}

// Lock locks m. In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *_RedisDMutex) Lock(ctx context.Context) error {
	return m.LockContext(ctx)
}

// Unlock unlocks m and returns the status of unlock.
func (m *_RedisDMutex) Unlock(ctx context.Context) (bool, error) {
	return m.UnlockContext(ctx)
}

// Extend resets the mutex's expiry and returns the status of expiry extension.
func (m *_RedisDMutex) Extend(ctx context.Context) (bool, error) {
	return m.ExtendContext(ctx)
}

// Valid returns true if the lock acquired through m is still valid. It may
// also return true erroneously if quorum is achieved during the call and at
// least one node then takes long enough to respond for the lock to expire.
func (m *_RedisDMutex) Valid(ctx context.Context) (bool, error) {
	return m.ValidContext(ctx)
}
