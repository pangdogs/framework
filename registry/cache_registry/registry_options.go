package cache_registry

import (
	"kit.golaxy.org/plugins/registry"
)

type Option struct{}

type RegistryOptions struct {
	Registry registry.Registry
}

type RegistryOption func(options *RegistryOptions)

func (Option) Default() RegistryOption {
	return func(options *RegistryOptions) {
		Option{}.Cached(nil)(options)
	}
}

func (Option) Cached(r registry.Registry) RegistryOption {
	return func(o *RegistryOptions) {
		o.Registry = r
	}
}
