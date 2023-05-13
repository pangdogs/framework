package zap

import (
	"go.uber.org/zap"
)

type Field int16

const (
	ServiceField Field = 1 << iota
	RuntimeField
)

type Options struct {
	ZapLogger     *zap.Logger
	CallerMaxSkip int8
	Fields        Field
}

type Option func(options *Options)

type WithOption struct{}

func (WithOption) Default() Option {
	return func(options *Options) {
		WithOption{}.ZapLogger(zap.NewExample())(options)
		WithOption{}.Fields(ServiceField | RuntimeField)(options)
		WithOption{}.CallerMaxSkip(3)(options)
	}
}

func (WithOption) ZapLogger(logger *zap.Logger) Option {
	return func(options *Options) {
		if logger == nil {
			panic("options.ZapLogger can't be assigned to nil")
		}
		options.ZapLogger = logger
	}
}

func (WithOption) Fields(fields Field) Option {
	return func(options *Options) {
		options.Fields = fields
	}
}

func (WithOption) CallerMaxSkip(skip int8) Option {
	return func(options *Options) {
		if skip < 0 {
			panic("options.CallerMaxSkip can't be set to a value less than 0")
		}
		options.CallerMaxSkip = skip
	}
}
