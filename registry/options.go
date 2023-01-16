package registry

import "time"

type RegisterOptions struct {
	TTL time.Duration
}

type RegisterOption func(o *RegisterOptions)

type WithRegisterOption struct{}

func (WithRegisterOption) Default() RegisterOption {
	return func(o *RegisterOptions) {
		o.TTL = 3 * time.Second
	}
}

func (WithRegisterOption) TTL(ttl time.Duration) RegisterOption {
	return func(o *RegisterOptions) {
		o.TTL = ttl
	}
}
