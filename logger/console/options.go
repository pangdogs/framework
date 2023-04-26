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

type ConsoleOptions struct {
	Development     bool
	Level           logger.Level
	Fields          Field
	Separator       string
	TimestampLayout string
	CallerFullName  bool
}

type ConsoleOption func(options *ConsoleOptions)

type WithConsoleOption struct{}

func (WithConsoleOption) Default() ConsoleOption {
	return func(options *ConsoleOptions) {
		WithConsoleOption{}.Development(false)
		WithConsoleOption{}.Level(logger.InfoLevel)(options)
		WithConsoleOption{}.Fields(ServiceField | RuntimeField | TimestampField | LevelField | CallerField)(options)
		WithConsoleOption{}.Separator(`|`)(options)
		WithConsoleOption{}.TimestampLayout(time.RFC3339Nano)(options)
		WithConsoleOption{}.CallerFullName(false)(options)
	}
}

func (WithConsoleOption) Development(b bool) ConsoleOption {
	return func(options *ConsoleOptions) {
		options.Development = b
	}
}

func (WithConsoleOption) Level(level logger.Level) ConsoleOption {
	return func(options *ConsoleOptions) {
		options.Level = level
	}
}

func (WithConsoleOption) Fields(fields Field) ConsoleOption {
	return func(options *ConsoleOptions) {
		options.Fields = fields
	}
}

func (WithConsoleOption) Separator(sp string) ConsoleOption {
	return func(options *ConsoleOptions) {
		options.Separator = sp
	}
}

func (WithConsoleOption) TimestampLayout(layout string) ConsoleOption {
	return func(options *ConsoleOptions) {
		options.TimestampLayout = layout
	}
}

func (WithConsoleOption) CallerFullName(b bool) ConsoleOption {
	return func(options *ConsoleOptions) {
		options.CallerFullName = b
	}
}
