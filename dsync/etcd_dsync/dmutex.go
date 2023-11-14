package etcd_dsync

import (
	"context"
	"errors"
	"fmt"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcd_concurrency "go.etcd.io/etcd/client/v3/concurrency"
	"kit.golaxy.org/plugins/dsync"
	"kit.golaxy.org/plugins/log"
	"math"
	"strconv"
	"strings"
	"time"
)

func (s *_DSync) newMutex(name string, options dsync.DMutexOptions) *_DMutex {
	if s.options.KeyPrefix != "" {
		name = s.options.KeyPrefix + name
	}

	log.Debugf(s.ctx, "new dsync mutex %q", name)

	return &_DMutex{
		dsync:         s,
		name:          name,
		expiry:        options.Expiry,
		tries:         options.Tries,
		driftFactor:   options.DriftFactor,
		timeoutFactor: options.TimeoutFactor,
	}
}

type _DMutex struct {
	dsync         *_DSync
	name          string
	expiry        time.Duration
	tries         int
	driftFactor   float64
	timeoutFactor float64
	session       *etcd_concurrency.Session
	mutex         *etcd_concurrency.Mutex
	until         time.Time
}

// Name returns mutex name.
func (m *_DMutex) Name() string {
	return strings.TrimPrefix(m.name, m.dsync.options.KeyPrefix)
}

// Value returns the current random value. The value will be empty until a lock is acquired (or Value option is used).
func (m *_DMutex) Value() string {
	if m.session == nil {
		return ""
	}
	return strconv.Itoa(int(m.session.Lease()))
}

// Until returns the time of validity of acquired lock. The value will be zero value until a lock is acquired.
func (m *_DMutex) Until() time.Time {
	return m.until
}

// Lock locks m. In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *_DMutex) Lock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	expirySec := math.Ceil(m.expiry.Seconds())

	session, err := etcd_concurrency.NewSession(m.dsync.client, etcd_concurrency.WithTTL(int(expirySec)))
	if err != nil {
		return fmt.Errorf("%w: %w", dsync.ErrDsync, err)
	}

	mutex := etcd_concurrency.NewMutex(session, m.name)

	start := time.Now()
	ctx, _ = context.WithTimeout(ctx, time.Duration((expirySec*m.timeoutFactor*float64(m.tries))*float64(time.Second)))

	if err = mutex.Lock(ctx); err != nil {
		return fmt.Errorf("%w: %w", dsync.ErrDsync, err)
	}

	if _, err = m.dsync.client.KeepAlive(ctx, session.Lease()); err != nil {
		mutex.Unlock(context.Background())
		return fmt.Errorf("%w: %w", dsync.ErrDsync, err)
	}

	m.clean()

	m.session = session
	m.mutex = mutex

	now := time.Now()
	m.until = now.Add(m.expiry - now.Sub(start) - time.Duration(int64(expirySec*m.driftFactor)))

	log.Debugf(m.dsync.ctx, "dsync mutex %q is locked", m.name)

	return nil
}

// Unlock unlocks m and returns the status of unlock.
func (m *_DMutex) Unlock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if m.mutex == nil {
		return dsync.ErrNotAcquired
	}

	if err := m.mutex.Unlock(ctx); err != nil {
		if errors.Is(err, rpctypes.ErrKeyNotFound) {
			m.clean()
			return dsync.ErrNotAcquired
		}
		return fmt.Errorf("%w: %w", dsync.ErrDsync, err)
	}

	m.clean()

	log.Debugf(m.dsync.ctx, "dsync mutex %q is unlocked", m.name)

	return nil
}

// Extend resets the mutex's expiry and returns the status of expiry extension.
func (m *_DMutex) Extend(ctx context.Context) error {
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
		return fmt.Errorf("%w: %w", dsync.ErrDsync, err)
	}

	log.Debugf(m.dsync.ctx, "dsync mutex %q is extended", m.name)

	return nil
}

// Valid returns true if the lock acquired through m is still valid. It may
// also return true erroneously if quorum is achieved during the call and at
// least one node then takes long enough to respond for the lock to expire.
func (m *_DMutex) Valid(ctx context.Context) (bool, error) {
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

func (m *_DMutex) clean() {
	if m.session != nil {
		m.session.Close()
	}
	m.session = nil
	m.mutex = nil
	m.until = time.Time{}
}
