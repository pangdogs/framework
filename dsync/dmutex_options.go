package dsync

import (
	"fmt"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/golaxy/util/uid"
	"math/rand"
	"time"
)

// Option is a helper struct to provide default options.
type Option struct{}

type (
	// A DelayFunc is used to decide the amount of time to wait between retries.
	DelayFunc = func(tries int) time.Duration
	// GenValueFunc is used to generate a random value.
	GenValueFunc = func() (string, error)
)

// DMutexOptions represents the options for acquiring a distributed mutex.
type DMutexOptions struct {
	Expiry        time.Duration
	Tries         int
	DelayFunc     DelayFunc
	DriftFactor   float64
	TimeoutFactor float64
	GenValueFunc  GenValueFunc
	Value         string
}

// DMutexOption represents a configuration option for acquiring a distributed mutex.
type DMutexOption func(options *DMutexOptions)

// Default sets the default options for acquiring a distributed mutex.
func (Option) Default() DMutexOption {
	defaultRetryDelayFunc := func(tries int) time.Duration {
		const (
			minRetryDelayMilliSec = 50
			maxRetryDelayMilliSec = 250
		)
		return time.Duration(rand.Intn(maxRetryDelayMilliSec-minRetryDelayMilliSec)+minRetryDelayMilliSec) * time.Millisecond
	}

	defaultGenValueFunc := func() (string, error) {
		return string(uid.New()), nil
	}

	return func(options *DMutexOptions) {
		Option{}.Expiry(8 * time.Second)(options)
		Option{}.Tries(32)(options)
		Option{}.RetryDelayFunc(defaultRetryDelayFunc)(options)
		Option{}.DriftFactor(0.01)(options)
		Option{}.TimeoutFactor(0.05)(options)
		Option{}.GenValueFunc(defaultGenValueFunc)(options)
		Option{}.Value("")(options)
	}
}

// Expiry can be used to set the expiry of a mutex to the given value.
func (Option) Expiry(expiry time.Duration) DMutexOption {
	return func(options *DMutexOptions) {
		options.Expiry = expiry
	}
}

// Tries can be used to set the number of times lock acquire is attempted.
func (Option) Tries(tries int) DMutexOption {
	return func(options *DMutexOptions) {
		options.Tries = tries
	}
}

// RetryDelay can be used to set the amount of time to wait between retries.
func (Option) RetryDelay(delay time.Duration) DMutexOption {
	return func(options *DMutexOptions) {
		options.DelayFunc = func(tries int) time.Duration {
			return delay
		}
	}
}

// RetryDelayFunc can be used to override default delay behavior.
func (Option) RetryDelayFunc(fn DelayFunc) DMutexOption {
	return func(options *DMutexOptions) {
		if fn == nil {
			panic(fmt.Errorf("%w: option DelayFunc can't be assigned to nil", golaxy.ErrArgs))
		}
		options.DelayFunc = fn
	}
}

// DriftFactor can be used to set the clock drift factor.
func (Option) DriftFactor(factor float64) DMutexOption {
	return func(options *DMutexOptions) {
		options.DriftFactor = factor
	}
}

// TimeoutFactor can be used to set the timeout factor.
func (Option) TimeoutFactor(factor float64) DMutexOption {
	return func(options *DMutexOptions) {
		options.TimeoutFactor = factor
	}
}

// GenValueFunc can be used to set the custom value generator.
func (Option) GenValueFunc(fn GenValueFunc) DMutexOption {
	return func(options *DMutexOptions) {
		if fn == nil {
			panic(fmt.Errorf("%w: option GenValueFunc can't be assigned to nil", golaxy.ErrArgs))
		}
		options.GenValueFunc = fn
	}
}

// Value can be used to assign the random value without having to call lock.
// This allows the ownership of a lock to be "transferred" and allows the lock to be unlocked from elsewhere.
func (Option) Value(v string) DMutexOption {
	return func(options *DMutexOptions) {
		options.Value = v
	}
}
