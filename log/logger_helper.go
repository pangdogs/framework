package log

import (
	"errors"
	"fmt"
	gx_plugin "kit.golaxy.org/golaxy/plugin"
	"os"
)

// Trace logs a message at TraceLevel, spaces are added between operands when neither is a string and a newline is appended.
func Trace(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Log(TraceLevel, v...)
	}
}

// Traceln logs a message at TraceLevel, spaces are always added between operands and a newline is appended.
func Traceln(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logln(TraceLevel, v...)
	}
}

// Tracef logs a formatted message at TraceLevel.
func Tracef(pluginResolver gx_plugin.PluginResolver, format string, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logf(TraceLevel, format, v...)
	}
}

// Debug logs a message at DebugLevel, spaces are added between operands when neither is a string and a newline is appended.
func Debug(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Log(DebugLevel, v...)
	}
}

// Debugln logs a message at DebugLevel, spaces are always added between operands and a newline is appended.
func Debugln(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logln(DebugLevel, v...)
	}
}

// Debugf logs a formatted message at DebugLevel.
func Debugf(pluginResolver gx_plugin.PluginResolver, format string, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logf(DebugLevel, format, v...)
	}
}

// Info logs a message at InfoLevel, spaces are added between operands when neither is a string and a newline is appended.
func Info(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Log(InfoLevel, v...)
	}
}

// Infoln logs a message at InfoLevel, spaces are always added between operands and a newline is appended.
func Infoln(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logln(InfoLevel, v...)
	}
}

// Infof logs a formatted message at InfoLevel.
func Infof(pluginResolver gx_plugin.PluginResolver, format string, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logf(InfoLevel, format, v...)
	}
}

// Warn logs a message at WarnLevel, spaces are added between operands when neither is a string and a newline is appended.
func Warn(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Log(WarnLevel, v...)
	}
}

// Warnln logs a message at WarnLevel, spaces are always added between operands and a newline is appended.
func Warnln(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logln(WarnLevel, v...)
	}
}

// Warnf logs a formatted message at WarnLevel.
func Warnf(pluginResolver gx_plugin.PluginResolver, format string, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logf(WarnLevel, format, v...)
	}
}

// Error logs a message at ErrorLevel, spaces are added between operands when neither is a string and a newline is appended.
func Error(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Log(ErrorLevel, v...)
	}
}

// Errorln logs a message at ErrorLevel, spaces are always added between operands and a newline is appended.
func Errorln(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logln(ErrorLevel, v...)
	}
}

// Errorf logs a formatted message at ErrorLevel.
func Errorf(pluginResolver gx_plugin.PluginResolver, format string, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logf(ErrorLevel, format, v...)
	}
}

// DPanic logs a message at DPanicLevel, spaces are added between operands when neither is a string and a newline is appended.
func DPanic(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Log(DPanicLevel, v...)
	} else {
		panic(errors.New(fmt.Sprint(v...)))
	}
}

// DPanicln logs a message at DPanicLevel, spaces are always added between operands and a newline is appended.
func DPanicln(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logln(DPanicLevel, v...)
	} else {
		panic(errors.New(fmt.Sprintln(v...)))
	}
}

// DPanicf logs a formatted message at DPanicLevel.
func DPanicf(pluginResolver gx_plugin.PluginResolver, format string, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logf(DPanicLevel, format, v...)
	} else {
		panic(fmt.Errorf(format, v...))
	}
}

// Panic logs a message at PanicLevel, spaces are added between operands when neither is a string and a newline is appended.
func Panic(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Log(PanicLevel, v...)
	} else {
		panic(errors.New(fmt.Sprint(v...)))
	}
}

// Panicln logs a message at PanicLevel, spaces are always added between operands and a newline is appended.
func Panicln(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logln(PanicLevel, v...)
	} else {
		panic(errors.New(fmt.Sprintln(v...)))
	}
}

// Panicf logs a formatted message at PanicLevel.
func Panicf(pluginResolver gx_plugin.PluginResolver, format string, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logf(PanicLevel, format, v...)
	} else {
		panic(fmt.Errorf(format, v...))
	}
}

// Fatal logs a message at FatalLevel, spaces are added between operands when neither is a string and a newline is appended.
func Fatal(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Log(FatalLevel, v...)
	} else {
		os.Exit(1)
	}
}

// Fatalln logs a message at FatalLevel, spaces are always added between operands and a newline is appended.
func Fatalln(pluginResolver gx_plugin.PluginResolver, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logln(FatalLevel, v...)
	} else {
		os.Exit(1)
	}
}

// Fatalf logs a formatted message at FatalLevel.
func Fatalf(pluginResolver gx_plugin.PluginResolver, format string, v ...interface{}) {
	if log, err := gx_plugin.Using[Logger](pluginResolver, plugin.Name); err == nil {
		log.Logf(FatalLevel, format, v...)
	} else {
		os.Exit(1)
	}
}
