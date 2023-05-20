package dsync

import (
	"kit.golaxy.org/golaxy/uid"
	"math/rand"
	"time"
)

// A DelayFunc is used to decide the amount of time to wait between retries.
type DelayFunc = func(tries int) time.Duration

// GenValueFunc is used to generate a random value.
type GenValueFunc = func() (string, error)

// Options represents the options for acquiring a distributed mutex.
type Options struct {
	Expiry        time.Duration
	Tries         int
	DelayFunc     DelayFunc
	DriftFactor   float64
	TimeoutFactor float64
	GenValueFunc  GenValueFunc
	Value         string
}

// Option represents a configuration option for acquiring a distributed mutex.
type Option func(options *Options)

// WithOption is a helper struct to provide default options.
type WithOption struct{}

// Default sets the default options for acquiring a distributed mutex.
func (WithOption) Default() Option {
	defaultRetryDelayFunc := func(tries int) time.Duration {
		const (
			minRetryDelayMilliSec = 50
			maxRetryDelayMilliSec = 250
		)
		return time.Duration(rand.Intn(maxRetryDelayMilliSec-minRetryDelayMilliSec)+minRetryDelayMilliSec) * time.Millisecond
	}

	defaultGenValueFunc := func() (string, error) {
		return uid.New().String(), nil
	}

	return func(options *Options) {
		WithOption{}.Expiry(8 * time.Second)(options)
		WithOption{}.Tries(32)(options)
		WithOption{}.RetryDelayFunc(defaultRetryDelayFunc)(options)
		WithOption{}.DriftFactor(0.01)(options)
		WithOption{}.TimeoutFactor(0.05)(options)
		WithOption{}.GenValueFunc(defaultGenValueFunc)(options)
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
func (WithOption) RetryDelayFunc(fn DelayFunc) Option {
	return func(options *Options) {
		if fn == nil {
			panic("option DelayFunc can't be assigned to nil")
		}
		options.DelayFunc = fn
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
func (WithOption) GenValueFunc(fn GenValueFunc) Option {
	return func(options *Options) {
		if fn == nil {
			panic("option GenValueFunc can't be assigned to nil")
		}
		options.GenValueFunc = fn
	}
}

// Value can be used to assign the random value without having to call lock.
// This allows the ownership of a lock to be "transferred" and allows the lock to be unlocked from elsewhere.
func (WithOption) Value(v string) Option {
	return func(options *Options) {
		options.Value = v
	}
}
