package log

import (
	"fmt"
	"git.golaxy.org/core/plugin"
	"os"
)

// Trace logs a message at TraceLevel, spaces are added between operands when neither is a string and a newline is appended.
func Trace(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Log(TraceLevel, v...)
	}
}

// Traceln logs a message at TraceLevel, spaces are always added between operands and a newline is appended.
func Traceln(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logln(TraceLevel, v...)
	}
}

// Tracef logs a formatted message at TraceLevel.
func Tracef(provider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logf(TraceLevel, format, v...)
	}
}

// Debug logs a message at DebugLevel, spaces are added between operands when neither is a string and a newline is appended.
func Debug(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Log(DebugLevel, v...)
	}
}

// Debugln logs a message at DebugLevel, spaces are always added between operands and a newline is appended.
func Debugln(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logln(DebugLevel, v...)
	}
}

// Debugf logs a formatted message at DebugLevel.
func Debugf(provider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logf(DebugLevel, format, v...)
	}
}

// Info logs a message at InfoLevel, spaces are added between operands when neither is a string and a newline is appended.
func Info(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Log(InfoLevel, v...)
	}
}

// Infoln logs a message at InfoLevel, spaces are always added between operands and a newline is appended.
func Infoln(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logln(InfoLevel, v...)
	}
}

// Infof logs a formatted message at InfoLevel.
func Infof(provider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logf(InfoLevel, format, v...)
	}
}

// Warn logs a message at WarnLevel, spaces are added between operands when neither is a string and a newline is appended.
func Warn(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Log(WarnLevel, v...)
	}
}

// Warnln logs a message at WarnLevel, spaces are always added between operands and a newline is appended.
func Warnln(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logln(WarnLevel, v...)
	}
}

// Warnf logs a formatted message at WarnLevel.
func Warnf(provider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logf(WarnLevel, format, v...)
	}
}

// Error logs a message at ErrorLevel, spaces are added between operands when neither is a string and a newline is appended.
func Error(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Log(ErrorLevel, v...)
	}
}

// Errorln logs a message at ErrorLevel, spaces are always added between operands and a newline is appended.
func Errorln(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logln(ErrorLevel, v...)
	}
}

// Errorf logs a formatted message at ErrorLevel.
func Errorf(provider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logf(ErrorLevel, format, v...)
	}
}

// DPanic logs a message at DPanicLevel, spaces are added between operands when neither is a string and a newline is appended.
func DPanic(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Log(DPanicLevel, v...)
	}
}

// DPanicln logs a message at DPanicLevel, spaces are always added between operands and a newline is appended.
func DPanicln(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logln(DPanicLevel, v...)
	}
}

// DPanicf logs a formatted message at DPanicLevel.
func DPanicf(provider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logf(DPanicLevel, format, v...)
	}
}

// Panic logs a message at PanicLevel, spaces are added between operands when neither is a string and a newline is appended.
func Panic(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Log(PanicLevel, v...)
	} else {
		panic(fmt.Sprint(v...))
	}
}

// Panicln logs a message at PanicLevel, spaces are always added between operands and a newline is appended.
func Panicln(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logln(PanicLevel, v...)
	} else {
		panic(fmt.Sprintln(v...))
	}
}

// Panicf logs a formatted message at PanicLevel.
func Panicf(provider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logf(PanicLevel, format, v...)
	} else {
		panic(fmt.Sprintf(format, v...))
	}
}

// Fatal logs a message at FatalLevel, spaces are added between operands when neither is a string and a newline is appended.
func Fatal(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Log(FatalLevel, v...)
	} else {
		os.Exit(1)
	}
}

// Fatalln logs a message at FatalLevel, spaces are always added between operands and a newline is appended.
func Fatalln(provider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logln(FatalLevel, v...)
	} else {
		os.Exit(1)
	}
}

// Fatalf logs a formatted message at FatalLevel.
func Fatalf(provider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(provider); logger != nil {
		logger.Logf(FatalLevel, format, v...)
	} else {
		os.Exit(1)
	}
}
