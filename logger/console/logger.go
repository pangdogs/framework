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
	helper := (level & logger.Level(logger.HelperFlag)) != 0

	level &= 0x0f
	if !level.Enabled(l.options.Level) {
		return
	}

	skip := 2
	if helper {
		skip = 3
	}

	l.logInfo(level, skip, fmt.Sprint(v...), "\n")
}

// Logln writes a log entry, spaces are always added between operands and a newline is appended
func (l *_ConsoleLogger) Logln(level logger.Level, v ...interface{}) {
	helper := (level & logger.Level(logger.HelperFlag)) != 0

	level &= 0x0f
	if !level.Enabled(l.options.Level) {
		return
	}

	skip := 2
	if helper {
		skip = 3
	}

	l.logInfo(level, skip, fmt.Sprintln(v...), "")
}

// Logf writes a formatted log entry
func (l *_ConsoleLogger) Logf(level logger.Level, format string, v ...interface{}) {
	helper := (level & logger.Level(logger.HelperFlag)) != 0

	level &= 0x0f
	if !level.Enabled(l.options.Level) {
		return
	}

	skip := 2
	if helper {
		skip = 3
	}

	l.logInfo(level, skip, fmt.Sprintf(format, v...), "\n")
}

func (l *_ConsoleLogger) logInfo(level logger.Level, skip int, info, endln string) {
	var writer io.Writer

	switch level {
	case logger.ErrorLevel:
		writer = os.Stderr
	default:
		writer = os.Stdout
	}

	_, file, line, ok := runtime.Caller(skip)
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

	var fields [12]any
	var count int32

	if l.options.Fields&ServiceField != 0 {
		fields[count] = l.serviceField
		count++
		fields[count] = l.options.Separator
		count++
	}

	if l.options.Fields&TimestampField != 0 {
		fields[count] = time.Now().Format(l.options.TimeLayout)
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
	case logger.PanicLevel:
		panic(info)
	case logger.FatalLevel:
		os.Exit(1)
	}
}
