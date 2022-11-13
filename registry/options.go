package registry

import "time"

type RegisterOptions struct {
	TTL time.Duration
}

type WithRegisterOption func(options *RegisterOptions)

var RegisterOption = _RegisterOption{}

type _RegisterOption struct{}

func (_RegisterOption) TTL(ttl time.Duration) WithRegisterOption {
	return func(options *RegisterOptions) {
		options.TTL = ttl
	}
}
