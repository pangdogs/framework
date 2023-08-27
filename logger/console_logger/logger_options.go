package console_logger

import (
	"kit.golaxy.org/plugins/logger"
	"time"
)

type Option struct{}

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

func (Option) Default() LoggerOption {
	return func(options *LoggerOptions) {
		Option{}.Development(false)
		Option{}.Level(logger.InfoLevel)(options)
		Option{}.Fields(ServiceField | RuntimeField | TimestampField | LevelField | CallerField)(options)
		Option{}.Separator(`|`)(options)
		Option{}.TimestampLayout(time.RFC3339Nano)(options)
		Option{}.CallerFullName(false)(options)
	}
}

func (Option) Development(b bool) LoggerOption {
	return func(options *LoggerOptions) {
		options.Development = b
	}
}

func (Option) Level(level logger.Level) LoggerOption {
	return func(options *LoggerOptions) {
		options.Level = level
	}
}

func (Option) Fields(fields Field) LoggerOption {
	return func(options *LoggerOptions) {
		options.Fields = fields
	}
}

func (Option) Separator(sp string) LoggerOption {
	return func(options *LoggerOptions) {
		options.Separator = sp
	}
}

func (Option) TimestampLayout(layout string) LoggerOption {
	return func(options *LoggerOptions) {
		options.TimestampLayout = layout
	}
}

func (Option) CallerFullName(b bool) LoggerOption {
	return func(options *LoggerOptions) {
		options.CallerFullName = b
	}
}
