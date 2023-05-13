package console

import (
	"kit.golaxy.org/plugins/logger"
	"time"
)

type Field int16

const (
	ServiceField Field = 1 << iota
	RuntimeField
	TimestampField
	LevelField
	CallerField
)

type Options struct {
	Development     bool
	Level           logger.Level
	Fields          Field
	Separator       string
	TimestampLayout string
	CallerFullName  bool
}

type Option func(options *Options)

type WithOption struct{}

func (WithOption) Default() Option {
	return func(options *Options) {
		WithOption{}.Development(false)
		WithOption{}.Level(logger.InfoLevel)(options)
		WithOption{}.Fields(ServiceField | RuntimeField | TimestampField | LevelField | CallerField)(options)
		WithOption{}.Separator(`|`)(options)
		WithOption{}.TimestampLayout(time.RFC3339Nano)(options)
		WithOption{}.CallerFullName(false)(options)
	}
}

func (WithOption) Development(b bool) Option {
	return func(options *Options) {
		options.Development = b
	}
}

func (WithOption) Level(level logger.Level) Option {
	return func(options *Options) {
		options.Level = level
	}
}

func (WithOption) Fields(fields Field) Option {
	return func(options *Options) {
		options.Fields = fields
	}
}

func (WithOption) Separator(sp string) Option {
	return func(options *Options) {
		options.Separator = sp
	}
}

func (WithOption) TimestampLayout(layout string) Option {
	return func(options *Options) {
		options.TimestampLayout = layout
	}
}

func (WithOption) CallerFullName(b bool) Option {
	return func(options *Options) {
		options.CallerFullName = b
	}
}
