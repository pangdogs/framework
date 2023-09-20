package etcd_dsync

import (
	"context"
	"errors"
	"fmt"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	etcd_concurrency "go.etcd.io/etcd/client/v3/concurrency"
	"kit.golaxy.org/plugins/dsync"
	"kit.golaxy.org/plugins/logger"
	"math"
	"strconv"
	"strings"
	"time"
)

func newEtcdMutex(es *_EtcdDSync, name string, options dsync.DMutexOptions) dsync.DMutex {
	if es.options.KeyPrefix != "" {
		name = es.options.KeyPrefix + name
	}

	logger.Debugf(es.ctx, "new dsync mutex %q", name)

	return &_EtcdDMutex{
		es:      es,
		name:    name,
		options: options,
	}
}

type _EtcdDMutex struct {
	es      *_EtcdDSync
	name    string
	options dsync.DMutexOptions
	session *etcd_concurrency.Session
	mutex   *etcd_concurrency.Mutex
	until   time.Time
}

// Name returns mutex name.
func (m *_EtcdDMutex) Name() string {
	return strings.TrimPrefix(m.name, m.es.options.KeyPrefix)
}

// Value returns the current random value. The value will be empty until a lock is acquired (or Value option is used).
func (m *_EtcdDMutex) Value() string {
	if m.session == nil {
		return ""
	}
	return strconv.Itoa(int(m.session.Lease()))
}

// Until returns the time of validity of acquired lock. The value will be zero value until a lock is acquired.
func (m *_EtcdDMutex) Until() time.Time {
	return m.until
}

// Lock locks m. In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *_EtcdDMutex) Lock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	expirySec := math.Ceil(m.options.Expiry.Seconds())

	session, err := etcd_concurrency.NewSession(m.es.client, etcd_concurrency.WithTTL(int(expirySec)))
	if err != nil {
		return err
	}

	mutex := etcd_concurrency.NewMutex(session, m.name)
	timeout := time.Duration((expirySec * m.options.TimeoutFactor * float64(m.options.Tries)) * float64(time.Second))
	start := time.Now()

	err = func() error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return mutex.Lock(ctx)
	}()
	if err != nil {
		return err
	}

	_, err = m.es.client.KeepAlive(ctx, session.Lease())
	if err != nil {
		mutex.Unlock(m.es.client.Ctx())
		return err
	}

	m.clean()

	now := time.Now()
	until := now.Add(m.options.Expiry - now.Sub(start) - time.Duration(int64(expirySec*m.options.DriftFactor)))

	m.session = session
	m.mutex = mutex
	m.until = until

	logger.Debugf(m.es.ctx, "dsync mutex %q is locked", m.name)

	return nil
}

// Unlock unlocks m and returns the status of unlock.
func (m *_EtcdDMutex) Unlock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if m.mutex == nil {
		return dsync.ErrNotAcquired
	}

	err := m.mutex.Unlock(ctx)
	if err != nil {
		if errors.Is(err, rpctypes.ErrKeyNotFound) {
			m.clean()
			return dsync.ErrNotAcquired
		}
		return err
	}

	logger.Debugf(m.es.ctx, "dsync mutex %q is unlocked", m.name)

	m.clean()
	return nil
}

// Extend resets the mutex's expiry and returns the status of expiry extension.
func (m *_EtcdDMutex) Extend(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if m.session == nil {
		return dsync.ErrNotAcquired
	}

	_, err := m.es.client.KeepAlive(ctx, m.session.Lease())
	if err != nil {
		if errors.Is(err, rpctypes.ErrLeaseNotFound) {
			return dsync.ErrNotAcquired
		}
		return err
	}

	logger.Debugf(m.es.ctx, "dsync mutex %q is extended", m.name)

	return nil
}

// Valid returns true if the lock acquired through m is still valid. It may
// also return true erroneously if quorum is achieved during the call and at
// least one node then takes long enough to respond for the lock to expire.
func (m *_EtcdDMutex) Valid(ctx context.Context) (bool, error) {
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

func (m *_EtcdDMutex) clean() {
	if m.session != nil {
		m.session.Close()
	}
	m.session = nil
	m.mutex = nil
	m.until = time.Time{}
}
