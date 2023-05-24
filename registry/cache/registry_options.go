package cache

import (
	"kit.golaxy.org/plugins/registry"
)

type WithOption struct{}

type RegistryOptions struct {
	Registry registry.Registry
}

type RegistryOption func(options *RegistryOptions)

func (WithOption) Default() RegistryOption {
	return func(options *RegistryOptions) {
		WithOption{}.Cached(nil)(options)
	}
}

func (WithOption) Cached(r registry.Registry) RegistryOption {
	return func(o *RegistryOptions) {
		o.Registry = r
	}
}
