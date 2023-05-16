package dsync

import (
	crand "crypto/rand"
	"encoding/base64"
	"math/rand"
	"time"
)

// A DelayFunc is used to decide the amount of time to wait between retries.
type DelayFunc func(tries int) time.Duration

// GenValueFunc is used to generate a random value.
type GenValueFunc func() (string, error)

type Options struct {
	Expiry        time.Duration
	Tries         int
	DelayFunc     DelayFunc
	DriftFactor   float64
	TimeoutFactor float64
	GenValueFunc  GenValueFunc
	Value         string
}

type Option func(options *Options)

type WithOption struct{}

func (WithOption) Default() Option {
	const (
		minRetryDelayMilliSec = 10
		maxRetryDelayMilliSec = 50
	)

	defaultRetryDelayFunc := func(tries int) time.Duration {
		return time.Duration(rand.Intn(maxRetryDelayMilliSec-minRetryDelayMilliSec)+minRetryDelayMilliSec) * time.Millisecond
	}

	defaultGenValueFunc := func() (string, error) {
		b := make([]byte, 16)
		_, err := crand.Read(b)
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(b), nil
	}

	return func(options *Options) {
		WithOption{}.Expiry(3 * time.Second)
		WithOption{}.Tries(3)
		WithOption{}.RetryDelayFunc(defaultRetryDelayFunc)
		WithOption{}.DriftFactor(0.01)
		WithOption{}.TimeoutFactor(0.05)
		WithOption{}.GenValueFunc(defaultGenValueFunc)
	}
}

// Expiry can be used to set the expiry of a mutex to the given value.
func (WithOption) Expiry(expiry time.Duration) Option {
	return func(options *Options) {
		options.Expiry = expiry
	}
}

// Tries can be used to set the number of times lock acquire is attempted.
func (WithOption) Tries(tries int) Option {
	return func(options *Options) {
		options.Tries = tries
	}
}

// RetryDelay can be used to set the amount of time to wait between retries.
func (WithOption) RetryDelay(delay time.Duration) Option {
	return func(options *Options) {
		options.DelayFunc = func(tries int) time.Duration {
			return delay
		}
	}
}

// RetryDelayFunc can be used to override default delay behavior.
func (WithOption) RetryDelayFunc(delayFunc DelayFunc) Option {
	return func(options *Options) {
		options.DelayFunc = delayFunc
	}
}

// DriftFactor can be used to set the clock drift factor.
func (WithOption) DriftFactor(factor float64) Option {
	return func(options *Options) {
		options.DriftFactor = factor
	}
}

// TimeoutFactor can be used to set the timeout factor.
func (WithOption) TimeoutFactor(factor float64) Option {
	return func(options *Options) {
		options.TimeoutFactor = factor
	}
}

// GenValueFunc can be used to set the custom value generator.
func (WithOption) GenValueFunc(genValueFunc GenValueFunc) Option {
	return func(options *Options) {
		options.GenValueFunc = genValueFunc
	}
}

// Value can be used to assign the random value without having to call lock.
// This allows the ownership of a lock to be "transferred" and allows the lock to be unlocked from elsewhere.
func (WithOption) Value(v string) Option {
	return func(options *Options) {
		options.Value = v
	}
}
