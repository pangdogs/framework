package logger

import "kit.golaxy.org/golaxy/service"

func Trace(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(TraceLevel|(Level(3)<<4), v...)
	}
}

func Tracef(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(TraceLevel|(Level(3)<<4), format, v...)
	}
}

func Debug(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(DebugLevel|(Level(3)<<4), v...)
	}
}

func Debugf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(DebugLevel|(Level(3)<<4), format, v...)
	}
}

func Info(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(InfoLevel|(Level(3)<<4), v...)
	}
}

func Infof(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(InfoLevel|(Level(3)<<4), format, v...)
	}
}

func Warn(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(WarnLevel|(Level(3)<<4), v...)
	}
}

func Warnf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(WarnLevel|(Level(3)<<4), format, v...)
	}
}

func Error(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(ErrorLevel|(Level(3)<<4), v...)
	}
}

func Errorf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(ErrorLevel|(Level(3)<<4), format, v...)
	}
}

func Panic(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(PanicLevel|(Level(3)<<4), v...)
	}
}

func Panicf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(PanicLevel|(Level(3)<<4), format, v...)
	}
}

func Fatal(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(FatalLevel|(Level(3)<<4), v...)
	}
}

func Fatalf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(FatalLevel|(Level(3)<<4), format, v...)
	}
}
