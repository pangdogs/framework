package zap_logger

import (
	"go.uber.org/zap"
	"kit.golaxy.org/golaxy/runtime"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/logger"
	"reflect"
)

func newZapLogger(options ...LoggerOption) logger.Logger {
	opts := LoggerOptions{}
	Option{}.Default()(&opts)

	for i := range options {
		options[i](&opts)
	}

	return &_ZapLogger{
		options: opts,
	}
}

type _ZapLogger struct {
	options        LoggerOptions
	sugaredLoggers []*zap.SugaredLogger
}

// InitSP 初始化服务插件
func (l *_ZapLogger) InitSP(ctx service.Context) {
	l.sugaredLoggers = make([]*zap.SugaredLogger, l.options.CallerMaxSkip)

	for i := range l.sugaredLoggers {
		options := []zap.Option{zap.AddCallerSkip(i)}
		if l.options.Fields&ServiceField != 0 {
			options = append(options, zap.Fields(zap.String("service", ctx.String())))
		}
		l.sugaredLoggers[i] = l.options.ZapLogger.WithOptions(options...).Sugar()
	}

	logger.Infof(ctx, "init service plugin %q with %q", definePlugin.Name, util.TypeOfAnyFullName(*l))
}

// ShutSP 关闭服务插件
func (l *_ZapLogger) ShutSP(ctx service.Context) {
	logger.Infof(ctx, "shut service plugin %q", definePlugin.Name)
}

// InitRP 初始化运行时插件
func (l *_ZapLogger) InitRP(ctx runtime.Context) {
	l.sugaredLoggers = make([]*zap.SugaredLogger, l.options.CallerMaxSkip)
	for i := range l.sugaredLoggers {
		options := []zap.Option{zap.AddCallerSkip(i)}
		if l.options.Fields&ServiceField != 0 {
			options = append(options, zap.Fields(zap.String("service", service.Current(ctx).String())))
		}
		if l.options.Fields&RuntimeField != 0 {
			options = append(options, zap.Fields(zap.String("runtime", ctx.String())))
		}
		l.sugaredLoggers[i] = l.options.ZapLogger.WithOptions(options...).Sugar()
	}

	logger.Infof(ctx, "init runtime plugin %q with %q", definePlugin.Name, reflect.TypeOf(_ZapLogger{}))
}

// ShutRP 关闭运行时插件
func (l *_ZapLogger) ShutRP(ctx runtime.Context) {
	logger.Infof(ctx, "shut runtime plugin %q", definePlugin.Name)
}

// Log writes a log entry, spaces are added between operands when neither is a string and a newline is appended
func (l *_ZapLogger) Log(level logger.Level, v ...interface{}) {
	level, skip := level.UnpackSkip()

	sugaredLogger := l.getSugaredLogger(skip + 1)
	if sugaredLogger == nil {
		return
	}

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
	case logger.DPanicLevel:
		sugaredLogger.DPanic(v...)
	case logger.PanicLevel:
		sugaredLogger.Panic(v...)
	case logger.FatalLevel:
		sugaredLogger.Fatal(v...)
	}
}

// Logln writes a log entry, spaces are always added between operands and a newline is appended
func (l *_ZapLogger) Logln(level logger.Level, v ...interface{}) {
	level, skip := level.UnpackSkip()

	sugaredLogger := l.getSugaredLogger(skip + 1)
	if sugaredLogger == nil {
		return
	}

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
	case logger.DPanicLevel:
		sugaredLogger.DPanicln(v...)
	case logger.PanicLevel:
		sugaredLogger.Panicln(v...)
	case logger.FatalLevel:
		sugaredLogger.Fatalln(v...)
	}
}

// Logf writes a formatted log entry
func (l *_ZapLogger) Logf(level logger.Level, format string, v ...interface{}) {
	level, skip := level.UnpackSkip()

	sugaredLogger := l.getSugaredLogger(skip + 1)
	if sugaredLogger == nil {
		return
	}

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
	case logger.DPanicLevel:
		sugaredLogger.DPanicf(format, v...)
	case logger.PanicLevel:
		sugaredLogger.Panicf(format, v...)
	case logger.FatalLevel:
		sugaredLogger.Fatalf(format, v...)
	}
}

func (l *_ZapLogger) getSugaredLogger(skip int8) *zap.SugaredLogger {
	if skip >= 0 && int(skip) < len(l.sugaredLoggers) {
		return l.sugaredLoggers[skip]
	}
	return nil
}
