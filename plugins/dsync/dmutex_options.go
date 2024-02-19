package dsync

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/core/util/uid"
	"math/rand"
	"time"
)

type (
	// A DelayFunc is used to decide the amount of time to wait between retries.
	DelayFunc = generic.Func1[int, time.Duration]
	// GenValueFunc is used to generate a random value.
	GenValueFunc = generic.PairFunc0[string, error]
)

// DistMutexOptions represents the options for acquiring a distributed mutex.
type DistMutexOptions struct {
	Expiry        time.Duration
	Tries         int
	DelayFunc     DelayFunc
	DriftFactor   float64
	TimeoutFactor float64
	GenValueFunc  GenValueFunc
	Value         string
}

var With _Option

type _Option struct{}

// Default sets the default options for acquiring a distributed mutex.
func (_Option) Default() option.Setting[DistMutexOptions] {
	defaultRetryDelayFunc := func(tries int) time.Duration {
		const (
			minRetryDelayMilliSec = 10
			maxRetryDelayMilliSec = 150
		)
		return time.Duration(rand.Intn(maxRetryDelayMilliSec-minRetryDelayMilliSec)+minRetryDelayMilliSec) * time.Millisecond
	}

	defaultGenValueFunc := func() (string, error) {
		return string(uid.New()), nil
	}

	return func(options *DistMutexOptions) {
		With.Expiry(3 * time.Second)(options)
		With.Tries(15)(options)
		With.RetryDelayFunc(defaultRetryDelayFunc)(options)
		With.DriftFactor(0.01)(options)
		With.TimeoutFactor(0.10)(options)
		With.GenValueFunc(defaultGenValueFunc)(options)
		With.Value("")(options)
	}
}

// Expiry can be used to set the expiry of a mutex to the given value.
func (_Option) Expiry(expiry time.Duration) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.Expiry = expiry
	}
}

// Tries can be used to set the number of times lock acquire is attempted.
func (_Option) Tries(tries int) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.Tries = tries
	}
}

// RetryDelay can be used to set the amount of time to wait between retries.
func (_Option) RetryDelay(delay time.Duration) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.DelayFunc = func(tries int) time.Duration {
			return delay
		}
	}
}

// RetryDelayFunc can be used to override default delay behavior.
func (_Option) RetryDelayFunc(fn DelayFunc) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		if fn == nil {
			panic(fmt.Errorf("%w: option DelayFunc can't be assigned to nil", core.ErrArgs))
		}
		options.DelayFunc = fn
	}
}

// DriftFactor can be used to set the clock drift factor.
func (_Option) DriftFactor(factor float64) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.DriftFactor = factor
	}
}

// TimeoutFactor can be used to set the timeout factor.
func (_Option) TimeoutFactor(factor float64) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.TimeoutFactor = factor
	}
}

// GenValueFunc can be used to set the custom value generator.
func (_Option) GenValueFunc(fn GenValueFunc) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		if fn == nil {
			panic(fmt.Errorf("%w: option GenValueFunc can't be assigned to nil", core.ErrArgs))
		}
		options.GenValueFunc = fn
	}
}

// Value can be used to assign the random value without having to call lock.
// This allows the ownership of a lock to be "transferred" and allows the lock to be unlocked from elsewhere.
func (_Option) Value(v string) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.Value = v
	}
}
