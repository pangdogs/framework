package zap

import (
	"go.uber.org/zap"
)

type ZapOptions struct {
	ZapLogger *zap.Logger
}

type ZapOption func(options *ZapOptions)

type WithZapOption struct{}

func (WithZapOption) Default() ZapOption {
	return func(options *ZapOptions) {
		WithZapOption{}.ZapLogger(nil)(options)
	}
}

func (WithZapOption) ZapLogger(v *zap.Logger) ZapOption {
	return func(options *ZapOptions) {
		options.ZapLogger = v
	}
}
