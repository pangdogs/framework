package zap_logger

import (
	"fmt"
	"go.uber.org/zap"
	"kit.golaxy.org/golaxy"
)

type Option struct{}

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

func (Option) Default() LoggerOption {
	return func(options *LoggerOptions) {
		Option{}.ZapLogger(zap.NewExample())(options)
		Option{}.Fields(ServiceField | RuntimeField)(options)
		Option{}.CallerMaxSkip(3)(options)
	}
}

func (Option) ZapLogger(logger *zap.Logger) LoggerOption {
	return func(options *LoggerOptions) {
		if logger == nil {
			panic(fmt.Errorf("%w: option ZapLogger can't be assigned to nil", golaxy.ErrArgs))
		}
		options.ZapLogger = logger
	}
}

func (Option) Fields(fields Field) LoggerOption {
	return func(options *LoggerOptions) {
		options.Fields = fields
	}
}

func (Option) CallerMaxSkip(skip int8) LoggerOption {
	return func(options *LoggerOptions) {
		if skip < 0 {
			panic(fmt.Errorf("%w: option CallerMaxSkip can't be set to a value less than 0", golaxy.ErrArgs))
		}
		options.CallerMaxSkip = skip
	}
}
