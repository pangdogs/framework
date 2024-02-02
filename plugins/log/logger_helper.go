package log

import (
	"fmt"
	"git.golaxy.org/core/plugin"
	"os"
)

// Trace logs a message at TraceLevel, spaces are added between operands when neither is a string and a newline is appended.
func Trace(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Log(TraceLevel, v...)
	}

}

// Traceln logs a message at TraceLevel, spaces are always added between operands and a newline is appended.
func Traceln(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logln(TraceLevel, v...)
	}
}

// Tracef logs a formatted message at TraceLevel.
func Tracef(pluginProvider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logf(TraceLevel, format, v...)
	}
}

// Debug logs a message at DebugLevel, spaces are added between operands when neither is a string and a newline is appended.
func Debug(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Log(DebugLevel, v...)
	}
}

// Debugln logs a message at DebugLevel, spaces are always added between operands and a newline is appended.
func Debugln(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logln(DebugLevel, v...)
	}
}

// Debugf logs a formatted message at DebugLevel.
func Debugf(pluginProvider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logf(DebugLevel, format, v...)
	}
}

// Info logs a message at InfoLevel, spaces are added between operands when neither is a string and a newline is appended.
func Info(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Log(InfoLevel, v...)
	}
}

// Infoln logs a message at InfoLevel, spaces are always added between operands and a newline is appended.
func Infoln(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logln(InfoLevel, v...)
	}
}

// Infof logs a formatted message at InfoLevel.
func Infof(pluginProvider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logf(InfoLevel, format, v...)
	}
}

// Warn logs a message at WarnLevel, spaces are added between operands when neither is a string and a newline is appended.
func Warn(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Log(WarnLevel, v...)
	}
}

// Warnln logs a message at WarnLevel, spaces are always added between operands and a newline is appended.
func Warnln(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logln(WarnLevel, v...)
	}
}

// Warnf logs a formatted message at WarnLevel.
func Warnf(pluginProvider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logf(WarnLevel, format, v...)
	}
}

// Error logs a message at ErrorLevel, spaces are added between operands when neither is a string and a newline is appended.
func Error(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Log(ErrorLevel, v...)
	}
}

// Errorln logs a message at ErrorLevel, spaces are always added between operands and a newline is appended.
func Errorln(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logln(ErrorLevel, v...)
	}
}

// Errorf logs a formatted message at ErrorLevel.
func Errorf(pluginProvider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logf(ErrorLevel, format, v...)
	}
}

// DPanic logs a message at DPanicLevel, spaces are added between operands when neither is a string and a newline is appended.
func DPanic(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Log(DPanicLevel, v...)
	}
}

// DPanicln logs a message at DPanicLevel, spaces are always added between operands and a newline is appended.
func DPanicln(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logln(DPanicLevel, v...)
	}
}

// DPanicf logs a formatted message at DPanicLevel.
func DPanicf(pluginProvider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logf(DPanicLevel, format, v...)
	}
}

// Panic logs a message at PanicLevel, spaces are added between operands when neither is a string and a newline is appended.
func Panic(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Log(PanicLevel, v...)
	} else {
		panic(fmt.Sprint(v...))
	}
}

// Panicln logs a message at PanicLevel, spaces are always added between operands and a newline is appended.
func Panicln(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logln(PanicLevel, v...)
	} else {
		panic(fmt.Sprintln(v...))
	}
}

// Panicf logs a formatted message at PanicLevel.
func Panicf(pluginProvider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logf(PanicLevel, format, v...)
	} else {
		panic(fmt.Sprintf(format, v...))
	}
}

// Fatal logs a message at FatalLevel, spaces are added between operands when neither is a string and a newline is appended.
func Fatal(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Log(FatalLevel, v...)
	} else {
		os.Exit(1)
	}
}

// Fatalln logs a message at FatalLevel, spaces are always added between operands and a newline is appended.
func Fatalln(pluginProvider plugin.PluginProvider, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logln(FatalLevel, v...)
	} else {
		os.Exit(1)
	}
}

// Fatalf logs a formatted message at FatalLevel.
func Fatalf(pluginProvider plugin.PluginProvider, format string, v ...interface{}) {
	if logger := Using(pluginProvider); logger != nil {
		logger.Logf(FatalLevel, format, v...)
	} else {
		os.Exit(1)
	}
}
