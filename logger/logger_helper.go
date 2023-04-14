package logger

import "kit.golaxy.org/golaxy/service"

func Trace(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(TraceLevel|Level(HelperFlag), v...)
	}
}

func Tracef(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(TraceLevel|Level(HelperFlag), format, v...)
	}
}

func Debug(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(DebugLevel|Level(HelperFlag), v...)
	}
}

func Debugf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(DebugLevel|Level(HelperFlag), format, v...)
	}
}

func Info(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(InfoLevel|Level(HelperFlag), v...)
	}
}

func Infof(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(InfoLevel|Level(HelperFlag), format, v...)
	}
}

func Warn(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(WarnLevel|Level(HelperFlag), v...)
	}
}

func Warnf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(WarnLevel|Level(HelperFlag), format, v...)
	}
}

func Error(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(ErrorLevel|Level(HelperFlag), v...)
	}
}

func Errorf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(ErrorLevel|Level(HelperFlag), format, v...)
	}
}

func Panic(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(PanicLevel|Level(HelperFlag), v...)
	}
}

func Panicf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(PanicLevel|Level(HelperFlag), format, v...)
	}
}

func Fatal(ctx service.Context, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Log(FatalLevel|Level(HelperFlag), v...)
	}
}

func Fatalf(ctx service.Context, format string, v ...interface{}) {
	log, ok := TryGet(ctx)
	if ok {
		log.Logf(FatalLevel|Level(HelperFlag), format, v...)
	}
}
