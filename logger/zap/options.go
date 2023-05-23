package zap

import (
	"go.uber.org/zap"
)

type WithOption struct{}

type Field int16

const (
	ServiceField Field = 1 << iota
	RuntimeField
)

type LoggerOptions struct {
	ZapLogger     *zap.Logger
	CallerMaxSkip int8
	Fields        Field
}

type LoggerOption func(options *LoggerOptions)

func (WithOption) Default() LoggerOption {
	return func(options *LoggerOptions) {
		WithOption{}.ZapLogger(zap.NewExample())(options)
		WithOption{}.Fields(ServiceField | RuntimeField)(options)
		WithOption{}.CallerMaxSkip(3)(options)
	}
}

func (WithOption) ZapLogger(logger *zap.Logger) LoggerOption {
	return func(options *LoggerOptions) {
		if logger == nil {
			panic("option ZapLogger can't be assigned to nil")
		}
		options.ZapLogger = logger
	}
}

func (WithOption) Fields(fields Field) LoggerOption {
	return func(options *LoggerOptions) {
		options.Fields = fields
	}
}

func (WithOption) CallerMaxSkip(skip int8) LoggerOption {
	return func(options *LoggerOptions) {
		if skip < 0 {
			panic("option CallerMaxSkip can't be set to a value less than 0")
		}
		options.CallerMaxSkip = skip
	}
}
