package cache

import (
	"kit.golaxy.org/plugins/registry"
)

type Options struct {
	Registry registry.Registry
}

type Option func(options *Options)

type WithOption struct{}

func (WithOption) Default() Option {
	return func(options *Options) {
		WithOption{}.Cached(nil)(options)
	}
}

func (WithOption) Cached(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}
