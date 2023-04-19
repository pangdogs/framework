package logger

import "kit.golaxy.org/golaxy/service"

// Trace logs a message at TraceLevel
func Trace(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(TraceLevel.PackSkip(1), v...)
	}
}

// Traceln logs a message at TraceLevel
func Traceln(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logln(TraceLevel.PackSkip(1), v...)
	}
}

// Tracef logs a formatted message at TraceLevel
func Tracef(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(TraceLevel.PackSkip(1), format, v...)
	}
}

// Debug logs a message at DebugLevel
func Debug(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(DebugLevel.PackSkip(1), v...)
	}
}

// Debugln logs a message at DebugLevel
func Debugln(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logln(DebugLevel.PackSkip(1), v...)
	}
}

// Debugf logs a formatted message at DebugLevel
func Debugf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(DebugLevel.PackSkip(1), format, v...)
	}
}

// Info logs a message at InfoLevel
func Info(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(InfoLevel.PackSkip(1), v...)
	}
}

// Infoln logs a message at InfoLevel
func Infoln(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logln(InfoLevel.PackSkip(1), v...)
	}
}

// Infof logs a formatted message at InfoLevel
func Infof(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(InfoLevel.PackSkip(1), format, v...)
	}
}

// Warn logs a message at WarnLevel
func Warn(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(WarnLevel.PackSkip(1), v...)
	}
}

// Warnln logs a message at WarnLevel
func Warnln(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logln(WarnLevel.PackSkip(1), v...)
	}
}

// Warnf logs a formatted message at WarnLevel
func Warnf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(WarnLevel.PackSkip(1), format, v...)
	}
}

// Error logs a message at ErrorLevel
func Error(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(ErrorLevel.PackSkip(1), v...)
	}
}

// Errorln logs a message at ErrorLevel
func Errorln(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logln(ErrorLevel.PackSkip(1), v...)
	}
}

// Errorf logs a formatted message at ErrorLevel
func Errorf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(ErrorLevel.PackSkip(1), format, v...)
	}
}

// Panic logs a message at PanicLevel
func Panic(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(PanicLevel.PackSkip(1), v...)
	}
}

// Panicln logs a message at PanicLevel
func Panicln(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logln(PanicLevel.PackSkip(1), v...)
	}
}

// Panicf logs a formatted message at PanicLevel
func Panicf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(PanicLevel.PackSkip(1), format, v...)
	}
}

// Fatal logs a message at FatalLevel
func Fatal(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(FatalLevel.PackSkip(1), v...)
	}
}

// Fatalln logs a message at FatalLevel
func Fatalln(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logln(FatalLevel.PackSkip(1), v...)
	}
}

// Fatalf logs a formatted message at FatalLevel
func Fatalf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(FatalLevel.PackSkip(1), format, v...)
	}
}
