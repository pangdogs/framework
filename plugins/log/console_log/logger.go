package console_log

import (
	"fmt"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/plugins/log"
	"io"
	"os"
	goruntime "runtime"
	"strings"
	"time"
)

func newLogger(settings ...option.Setting[LoggerOptions]) log.ILogger {
	return &_Logger{
		options: option.Make(Option{}.Default(), settings...),
	}
}

type _Logger struct {
	options     LoggerOptions
	serviceInfo string
	runtimeInfo string
}

// InitSP init service plugin
func (l *_Logger) InitSP(ctx service.Context) {
	l.serviceInfo = ctx.String()
}

// InitRP init runtime plugin
func (l *_Logger) InitRP(ctx runtime.Context) {
	l.serviceInfo = service.Current(ctx).String()
	l.runtimeInfo = ctx.String()
}

// Log writes a log entry, spaces are added between operands when neither is a string and a newline is appended.
func (l *_Logger) Log(level log.Level, v ...interface{}) {
	if !l.options.Level.Enabled(level) {
		l.interrupt(level, fmt.Sprint(v...))
		return
	}

	msg := fmt.Sprint(v...)
	l.logMessage(level, l.options.CallerSkip, msg, "\n")
	l.interrupt(level, msg)
}

// Logln writes a log entry, spaces are always added between operands and a newline is appended.
func (l *_Logger) Logln(level log.Level, v ...interface{}) {
	if !l.options.Level.Enabled(level) {
		l.interrupt(level, fmt.Sprintln(v...))
		return
	}

	msg := fmt.Sprintln(v...)
	l.logMessage(level, l.options.CallerSkip, msg, "")
	l.interrupt(level, msg)
}

// Logf writes a formatted log entry.
func (l *_Logger) Logf(level log.Level, format string, v ...interface{}) {
	if !l.options.Level.Enabled(level) {
		l.interrupt(level, fmt.Sprintf(format, v...))
		return
	}

	msg := fmt.Sprintf(format, v...)
	l.logMessage(level, l.options.CallerSkip, msg, "\n")
	l.interrupt(level, msg)
}

func (l *_Logger) logMessage(level log.Level, skip int, msg, endln string) {
	var writer io.Writer

	switch level {
	case log.ErrorLevel, log.DPanicLevel, log.PanicLevel, log.FatalLevel:
		writer = os.Stderr
	default:
		writer = os.Stdout
	}

	var fields [16]any
	var count int32

	if l.serviceInfo != "" && l.options.ServiceInfo {
		fields[count] = l.serviceInfo
		count++
		fields[count] = l.options.Separator
		count++
	}

	if l.runtimeInfo != "" && l.options.RuntimeInfo {
		fields[count] = l.runtimeInfo
		count++
		fields[count] = l.options.Separator
		count++
	}

	{
		fields[count] = time.Now().Format(l.options.TimestampLayout)
		count++
		fields[count] = l.options.Separator
		count++
	}

	{
		fields[count] = level
		count++
		fields[count] = l.options.Separator
		count++
	}

	{
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

	fields[count] = msg
	count++
	fields[count] = endln
	count++

	fmt.Fprint(writer, fields[:count]...)
}

func (l *_Logger) interrupt(level log.Level, msg string) {
	switch level {
	case log.DPanicLevel:
		if l.options.Development {
			panic(msg)
		}
	case log.PanicLevel:
		panic(msg)
	case log.FatalLevel:
		os.Exit(1)
	}
}
