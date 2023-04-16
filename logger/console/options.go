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
	Level          logger.Level
	Fields         Field
	Separator      string
	TimeLayout     string
	FullCallerName bool
}

type ConsoleOption func(options *ConsoleOptions)

type WithConsoleOption struct{}

func (WithConsoleOption) Default() ConsoleOption {
	return func(options *ConsoleOptions) {
		WithConsoleOption{}.Level(logger.InfoLevel)(options)
		WithConsoleOption{}.Fields(ServiceField | TimestampField | LevelField | CallerField)(options)
		WithConsoleOption{}.Separator(`|`)(options)
		WithConsoleOption{}.TimeLayout(time.RFC3339Nano)(options)
		WithConsoleOption{}.FullCallerName(false)(options)
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

func (WithConsoleOption) TimeLayout(v string) ConsoleOption {
	return func(options *ConsoleOptions) {
		options.TimeLayout = v
	}
}

func (WithConsoleOption) FullCallerName(v bool) ConsoleOption {
	return func(options *ConsoleOptions) {
		options.FullCallerName = v
	}
}
