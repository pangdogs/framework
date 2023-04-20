package console

import (
	"kit.golaxy.org/plugins/logger"
	"time"
)

type Field int16

const (
	ServiceField Field = 1 << iota
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
		WithConsoleOption{}.Fields(ServiceField | TimestampField | LevelField | CallerField)(options)
		WithConsoleOption{}.Separator(`|`)(options)
		WithConsoleOption{}.TimestampLayout(time.RFC3339Nano)(options)
		WithConsoleOption{}.CallerFullName(false)(options)
	}
}

func (WithConsoleOption) Development(v bool) ConsoleOption {
	return func(options *ConsoleOptions) {
		options.Development = v
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

func (WithConsoleOption) Separator(v string) ConsoleOption {
	return func(options *ConsoleOptions) {
		options.Separator = v
	}
}

func (WithConsoleOption) TimestampLayout(v string) ConsoleOption {
	return func(options *ConsoleOptions) {
		options.TimestampLayout = v
	}
}

func (WithConsoleOption) CallerFullName(v bool) ConsoleOption {
	return func(options *ConsoleOptions) {
		options.CallerFullName = v
	}
}
