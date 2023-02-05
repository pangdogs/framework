package logger

import "kit.golaxy.org/golaxy/service"

func Trace(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(TraceLevel, v...)
	}
}

func Tracef(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(TraceLevel, format, v...)
	}
}

func Debug(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(DebugLevel, v...)
	}
}

func Debugf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(DebugLevel, format, v...)
	}
}

func Info(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(InfoLevel, v...)
	}
}

func Infof(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(InfoLevel, format, v...)
	}
}

func Warn(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(WarnLevel, v...)
	}
}

func Warnf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(WarnLevel, format, v...)
	}
}

func Error(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(ErrorLevel, v...)
	}
}

func Errorf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(ErrorLevel, format, v...)
	}
}

func Panic(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(PanicLevel, v...)
	}
}

func Panicf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(PanicLevel, format, v...)
	}
}

func Fatal(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(FatalLevel, v...)
	}
}

func Fatalf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(FatalLevel, format, v...)
	}
}
