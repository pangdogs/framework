package cache

import (
	"kit.golaxy.org/plugins/registry"
)

type CacheOptions struct {
	Registry registry.Registry
}

type CacheOption func(options *CacheOptions)

type WithCacheOption struct{}

func (WithCacheOption) Default() CacheOption {
	return func(options *CacheOptions) {
		WithCacheOption{}.CacheClient(nil)(options)
	}
}

func (WithCacheOption) CacheClient(r registry.Registry) CacheOption {
	return func(o *CacheOptions) {
		o.Registry = r
	}
}
