package console

import (
	"fmt"
	"io"
	"kit.golaxy.org/golaxy/runtime"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/logger"
	"os"
	"reflect"
	goruntime "runtime"
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
	serviceField string
	runtimeField string
}

// InitService init service plugin
func (l *_ConsoleLogger) InitService(ctx service.Context) {
	l.serviceField = ctx.String()

	logger.Infof(ctx, "init service plugin %s with %s", definePlugin.Name, reflect.TypeOf(_ConsoleLogger{}))
}

// ShutService shut service plugin
func (l *_ConsoleLogger) ShutService(ctx service.Context) {
	logger.Infof(ctx, "shut service plugin %s", definePlugin.Name)
}

// InitRuntime init runtime plugin
func (l *_ConsoleLogger) InitRuntime(ctx runtime.Context) {
	l.serviceField = service.Get(ctx).String()
	l.runtimeField = ctx.String()

	logger.Infof(ctx, "init runtime plugin %s with %s", definePlugin.Name, reflect.TypeOf(_ConsoleLogger{}))
}

// ShutRuntime shut runtime plugin
func (l *_ConsoleLogger) ShutRuntime(ctx runtime.Context) {
	logger.Infof(ctx, "shut runtime plugin %s", definePlugin.Name)
}

// Log writes a log entry, spaces are added between operands when neither is a string and a newline is appended.
func (l *_ConsoleLogger) Log(level logger.Level, v ...interface{}) {
	level, skip := level.UnpackSkip()

	if !l.options.Level.Enabled(level) {
		return
	}

	l.logInfo(level, skip+2, fmt.Sprint(v...), "\n")
}

// Logln writes a log entry, spaces are always added between operands and a newline is appended.
func (l *_ConsoleLogger) Logln(level logger.Level, v ...interface{}) {
	level, skip := level.UnpackSkip()

	if !l.options.Level.Enabled(level) {
		return
	}

	l.logInfo(level, skip+2, fmt.Sprintln(v...), "")
}

// Logf writes a formatted log entry.
func (l *_ConsoleLogger) Logf(level logger.Level, format string, v ...interface{}) {
	level, skip := level.UnpackSkip()

	if !l.options.Level.Enabled(level) {
		return
	}

	l.logInfo(level, skip+2, fmt.Sprintf(format, v...), "\n")
}

func (l *_ConsoleLogger) logInfo(level logger.Level, skip int8, info, endln string) {
	var writer io.Writer

	switch level {
	case logger.ErrorLevel, logger.DPanicLevel, logger.PanicLevel, logger.FatalLevel:
		writer = os.Stderr
	default:
		writer = os.Stdout
	}

	var fields [16]any
	var count int32

	if l.serviceField != "" && l.options.Fields&ServiceField != 0 {
		fields[count] = l.serviceField
		count++
		fields[count] = l.options.Separator
		count++
	}

	if l.runtimeField != "" && l.options.Fields&RuntimeField != 0 {
		fields[count] = l.runtimeField
		count++
		fields[count] = l.options.Separator
		count++
	}

	if l.options.Fields&TimestampField != 0 {
		fields[count] = time.Now().Format(l.options.TimestampLayout)
		count++
		fields[count] = l.options.Separator
		count++
	}

	if l.options.Fields&LevelField != 0 {
		fields[count] = level
		count++
		fields[count] = l.options.Separator
		count++
	}

	if l.options.Fields&CallerField != 0 {
		_, file, line, ok := goruntime.Caller(int(skip))
		if !ok {
			file = "???"
			line = 0
		} else {
			if !l.options.CallerFullName {
				idx := strings.LastIndexByte(file, '/')
				if idx > 0 {
					idx = strings.LastIndexByte(file[:idx], '/')
					if idx > 0 {
						file = file[idx+1:]
					}
				}
			}
		}

		fields[count] = file
		count++
		fields[count] = ":"
		count++
		fields[count] = line
		count++
		fields[count] = l.options.Separator
		count++
	}

	fields[count] = info
	count++
	fields[count] = endln
	count++

	fmt.Fprint(writer, fields[:count]...)

	switch level {
	case logger.DPanicLevel:
		if l.options.Development {
			panic(info)
		}
	case logger.PanicLevel:
		panic(info)
	case logger.FatalLevel:
		os.Exit(1)
	}
}
