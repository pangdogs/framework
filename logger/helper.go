package logger

import (
	"github.com/galaxy-kit/galaxy-go/service"
)

func Log(ctx service.Context, level Level, args ...interface{}) {
	logger, ok := Plugin.TryGet(ctx)
	if !ok {
		return
	}
	logger.Log(level, args...)
}

func Logf(ctx service.Context, level Level, format string, args ...interface{}) {
	logger, ok := Plugin.TryGet(ctx)
	if !ok {
		return
	}
	logger.Logf(level, format, args...)
}

func Info(ctx service.Context, args ...interface{}) {
	logger, ok := Plugin.TryGet(ctx)
	if !ok {
		return
	}
	logger.Log(InfoLevel, args...)
}

func Infof(ctx service.Context, format string, args ...interface{}) {
	logger, ok := Plugin.TryGet(ctx)
	if !ok {
		return
	}
	logger.Logf(InfoLevel, format, args...)
}

func Trace(ctx service.Context, args ...interface{}) {
	logger, ok := Plugin.TryGet(ctx)
	if !ok {
		return
	}
	logger.Log(TraceLevel, args...)
}

func Tracef(ctx service.Context, format string, args ...interface{}) {
	logger, ok := Plugin.TryGet(ctx)
	if !ok {
		return
	}
	logger.Logf(TraceLevel, format, args...)
}

func Debug(ctx service.Context, args ...interface{}) {
	logger, ok := Plugin.TryGet(ctx)
	if !ok {
		return
	}
	logger.Log(DebugLevel, args...)
}

func Debugf(ctx service.Context, format string, args ...interface{}) {
	logger, ok := Plugin.TryGet(ctx)
	if !ok {
		return
	}
	logger.Logf(DebugLevel, format, args...)
}

func Warn(ctx service.Context, args ...interface{}) {
	logger, ok := Plugin.TryGet(ctx)
	if !ok {
		return
	}
	logger.Log(WarnLevel, args...)
}

func Warnf(ctx service.Context, format string, args ...interface{}) {
	logger, ok := Plugin.TryGet(ctx)
	if !ok {
		return
	}
	logger.Logf(WarnLevel, format, args...)
}

func Error(ctx service.Context, args ...interface{}) {
	logger, ok := Plugin.TryGet(ctx)
	if !ok {
		return
	}
	logger.Log(ErrorLevel, args...)
}

func Errorf(ctx service.Context, format string, args ...interface{}) {
	logger, ok := Plugin.TryGet(ctx)
	if !ok {
		return
	}
	logger.Logf(ErrorLevel, format, args...)
}

func Fatal(ctx service.Context, args ...interface{}) {
	logger, ok := Plugin.TryGet(ctx)
	if !ok {
		return
	}
	logger.Log(FatalLevel, args...)
}

func Fatalf(ctx service.Context, format string, args ...interface{}) {
	logger, ok := Plugin.TryGet(ctx)
	if !ok {
		return
	}
	logger.Logf(FatalLevel, format, args...)
}
