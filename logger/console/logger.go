package console

import (
	"fmt"
	"io"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/logger"
	"os"
	"runtime"
	"strings"
	"time"
)

func newConsoleLogger(options ...ConsoleOption) logger.Logger {
	opts := ConsoleOptions{}
	WithConsoleOption{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_ConsoleLogger{
		options: opts,
	}
}

type _ConsoleLogger struct {
	options      ConsoleOptions
	serviceCtx   service.Context
	serviceField string
}

// Init 初始化
func (l *_ConsoleLogger) Init(ctx service.Context) {
	l.serviceCtx = ctx
	l.serviceField = l.serviceCtx.String()
}

// Log writes a log entry, spaces are added between operands when neither is a string and a newline is appended
func (l *_ConsoleLogger) Log(level logger.Level, v ...interface{}) {
	if !level.Enabled(l.options.Level) {
		return
	}

	l.logInfo(level, fmt.Sprint(v...), "\n")
}

// Logln writes a log entry, spaces are always added between operands and a newline is appended
func (l *_ConsoleLogger) Logln(level logger.Level, v ...interface{}) {
	if !level.Enabled(l.options.Level) {
		return
	}

	l.logInfo(level, fmt.Sprintln(v...), "")
}

// Logf writes a formatted log entry
func (l *_ConsoleLogger) Logf(level logger.Level, format string, v ...interface{}) {
	if !level.Enabled(l.options.Level) {
		return
	}

	l.logInfo(level, fmt.Sprintf(format, v...), "\n")
}

func (l *_ConsoleLogger) logInfo(level logger.Level, info, endln string) {
	var writer io.Writer

	switch level {
	case logger.ErrorLevel:
		writer = os.Stderr
	default:
		writer = os.Stdout
	}

	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	} else {
		if !l.options.FullCallerName {
			idx := strings.LastIndexByte(file, '/')
			if idx > 0 {
				idx = strings.LastIndexByte(file[:idx], '/')
				if idx > 0 {
					file = file[idx+1:]
				}
			}
		}
	}

	fmt.Fprint(writer, l.serviceField, l.options.Separator, time.Now().Format(l.options.TimeLayout), l.options.Separator, level, l.options.Separator, file, ":", line, l.options.Separator, info, endln)

	switch level {
	case logger.PanicLevel:
		panic(info)
	case logger.FatalLevel:
		os.Exit(1)
	}
}
