package zap

import (
	"go.uber.org/zap"
)

type Field int16

const (
	ServiceField Field = 1 << iota
	RuntimeField
)

type ZapOptions struct {
	ZapLogger     *zap.Logger
	CallerMaxSkip int8
	Fields        Field
}

type ZapOption func(options *ZapOptions)

type WithZapOption struct{}

func (WithZapOption) Default() ZapOption {
	return func(options *ZapOptions) {
		WithZapOption{}.ZapLogger(zap.NewExample())(options)
		WithZapOption{}.Fields(ServiceField | RuntimeField)(options)
		WithZapOption{}.CallerMaxSkip(3)(options)
	}
}

func (WithZapOption) ZapLogger(logger *zap.Logger) ZapOption {
	return func(options *ZapOptions) {
		if logger == nil {
			panic("options.ZapLogger can't be assigned to nil")
		}
		options.ZapLogger = logger
	}
}

func (WithZapOption) Fields(fields Field) ZapOption {
	return func(options *ZapOptions) {
		options.Fields = fields
	}
}

func (WithZapOption) CallerMaxSkip(skip int8) ZapOption {
	return func(options *ZapOptions) {
		if skip < 0 {
			panic("options.CallerMaxSkip can't be set to a value less than 0")
		}
		options.CallerMaxSkip = skip
	}
}
