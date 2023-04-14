package logger

import "kit.golaxy.org/golaxy/service"

//go:inline
func Trace(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(TraceLevel, v...)
	}
}

//go:inline
func Tracef(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(TraceLevel, format, v...)
	}
}

//go:inline
func Debug(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(DebugLevel, v...)
	}
}

//go:inline
func Debugf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(DebugLevel, format, v...)
	}
}

//go:inline
func Info(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(InfoLevel, v...)
	}
}

//go:inline
func Infof(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(InfoLevel, format, v...)
	}
}

//go:inline
func Warn(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(WarnLevel, v...)
	}
}

//go:inline
func Warnf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(WarnLevel, format, v...)
	}
}

//go:inline
func Error(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(ErrorLevel, v...)
	}
}

//go:inline
func Errorf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(ErrorLevel, format, v...)
	}
}

//go:inline
func Panic(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(PanicLevel, v...)
	}
}

//go:inline
func Panicf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(PanicLevel, format, v...)
	}
}

//go:inline
func Fatal(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(FatalLevel, v...)
	}
}

//go:inline
func Fatalf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(FatalLevel, format, v...)
	}
}
