package console

import (
	"kit.golaxy.org/plugins/logger"
	"time"
)

type WithOption struct{}

type Field int16

const (
	ServiceField Field = 1 << iota
	RuntimeField
	TimestampField
	LevelField
	CallerField
)

type LoggerOptions struct {
	Development     bool
	Level           logger.Level
	Fields          Field
	Separator       string
	TimestampLayout string
	CallerFullName  bool
}

type LoggerOption func(options *LoggerOptions)

func (WithOption) Default() LoggerOption {
	return func(options *LoggerOptions) {
		WithOption{}.Development(false)
		WithOption{}.Level(logger.InfoLevel)(options)
		WithOption{}.Fields(ServiceField | RuntimeField | TimestampField | LevelField | CallerField)(options)
		WithOption{}.Separator(`|`)(options)
		WithOption{}.TimestampLayout(time.RFC3339Nano)(options)
		WithOption{}.CallerFullName(false)(options)
	}
}

func (WithOption) Development(b bool) LoggerOption {
	return func(options *LoggerOptions) {
		options.Development = b
	}
}

func (WithOption) Level(level logger.Level) LoggerOption {
	return func(options *LoggerOptions) {
		options.Level = level
	}
}

func (WithOption) Fields(fields Field) LoggerOption {
	return func(options *LoggerOptions) {
		options.Fields = fields
	}
}

func (WithOption) Separator(sp string) LoggerOption {
	return func(options *LoggerOptions) {
		options.Separator = sp
	}
}

func (WithOption) TimestampLayout(layout string) LoggerOption {
	return func(options *LoggerOptions) {
		options.TimestampLayout = layout
	}
}

func (WithOption) CallerFullName(b bool) LoggerOption {
	return func(options *LoggerOptions) {
		options.CallerFullName = b
	}
}
