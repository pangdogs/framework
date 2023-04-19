package zap

import (
	"go.uber.org/zap"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/logger"
	"reflect"
)

func newZapLogger(options ...ZapOption) logger.Logger {
	opts := ZapOptions{}
	WithZapOption{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_ZapLogger{
		options: opts,
	}
}

type _ZapLogger struct {
	options        ZapOptions
	serviceCtx     service.Context
	sugaredLoggers [10]*zap.SugaredLogger
}

// Init 初始化
func (l *_ZapLogger) Init(ctx service.Context) {
	l.serviceCtx = ctx

	if l.options.ZapLogger == nil {
		panic("option ZapLogger is nil, must be set")
	}

	for i := range l.sugaredLoggers {
		l.sugaredLoggers[i] = l.options.ZapLogger.WithOptions(zap.AddCallerSkip(i)).Sugar()
	}

	logger.Infof(ctx, "init plugin %s with %s", plugin.Name, reflect.TypeOf(_ZapLogger{}))
}

// Shut 关闭
func (l *_ZapLogger) Shut() {
	logger.Infof(l.serviceCtx, "shut plugin %s", plugin.Name)
}

// Log writes a log entry, spaces are added between operands when neither is a string and a newline is appended
func (l *_ZapLogger) Log(level logger.Level, v ...interface{}) {
	level, skip := level.UnpackSkip()
	skip += 1

	if skip < 0 || int(skip) >= len(l.sugaredLoggers) {
		return
	}

	sugaredLogger := l.sugaredLoggers[skip]

	switch level {
	case logger.TraceLevel:
		sugaredLogger.Debug(v...)
	case logger.DebugLevel:
		sugaredLogger.Debug(v...)
	case logger.InfoLevel:
		sugaredLogger.Info(v...)
	case logger.WarnLevel:
		sugaredLogger.Warn(v...)
	case logger.ErrorLevel:
		sugaredLogger.Error(v...)
	case logger.PanicLevel:
		sugaredLogger.Panic(v...)
	case logger.FatalLevel:
		sugaredLogger.Fatal(v...)
	}
}

// Logln writes a log entry, spaces are always added between operands and a newline is appended
func (l *_ZapLogger) Logln(level logger.Level, v ...interface{}) {
	level, skip := level.UnpackSkip()
	skip += 1

	if skip < 0 || int(skip) >= len(l.sugaredLoggers) {
		return
	}

	sugaredLogger := l.sugaredLoggers[skip]

	switch level {
	case logger.TraceLevel:
		sugaredLogger.Debugln(v...)
	case logger.DebugLevel:
		sugaredLogger.Debugln(v...)
	case logger.InfoLevel:
		sugaredLogger.Infoln(v...)
	case logger.WarnLevel:
		sugaredLogger.Warnln(v...)
	case logger.ErrorLevel:
		sugaredLogger.Errorln(v...)
	case logger.PanicLevel:
		sugaredLogger.Panicln(v...)
	case logger.FatalLevel:
		sugaredLogger.Fatalln(v...)
	}
}

// Logf writes a formatted log entry
func (l *_ZapLogger) Logf(level logger.Level, format string, v ...interface{}) {
	level, skip := level.UnpackSkip()
	skip += 1

	if skip < 0 || int(skip) >= len(l.sugaredLoggers) {
		return
	}

	sugaredLogger := l.sugaredLoggers[skip]

	switch level {
	case logger.TraceLevel:
		sugaredLogger.Debugf(format, v...)
	case logger.DebugLevel:
		sugaredLogger.Debugf(format, v...)
	case logger.InfoLevel:
		sugaredLogger.Infof(format, v...)
	case logger.WarnLevel:
		sugaredLogger.Warnf(format, v...)
	case logger.ErrorLevel:
		sugaredLogger.Errorf(format, v...)
	case logger.PanicLevel:
		sugaredLogger.Panicf(format, v...)
	case logger.FatalLevel:
		sugaredLogger.Fatalf(format, v...)
	}
}
