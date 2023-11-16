package log

import (
	"errors"
	"fmt"
	gxplugin "kit.golaxy.org/golaxy/plugin"
	"os"
)

// Trace logs a message at TraceLevel, spaces are added between operands when neither is a string and a newline is appended.
func Trace(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Log(TraceLevel, v...)
	}
}

// Traceln logs a message at TraceLevel, spaces are always added between operands and a newline is appended.
func Traceln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logln(TraceLevel, v...)
	}
}

// Tracef logs a formatted message at TraceLevel.
func Tracef(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logf(TraceLevel, format, v...)
	}
}

// Debug logs a message at DebugLevel, spaces are added between operands when neither is a string and a newline is appended.
func Debug(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Log(DebugLevel, v...)
	}
}

// Debugln logs a message at DebugLevel, spaces are always added between operands and a newline is appended.
func Debugln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logln(DebugLevel, v...)
	}
}

// Debugf logs a formatted message at DebugLevel.
func Debugf(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logf(DebugLevel, format, v...)
	}
}

// Info logs a message at InfoLevel, spaces are added between operands when neither is a string and a newline is appended.
func Info(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Log(InfoLevel, v...)
	}
}

// Infoln logs a message at InfoLevel, spaces are always added between operands and a newline is appended.
func Infoln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logln(InfoLevel, v...)
	}
}

// Infof logs a formatted message at InfoLevel.
func Infof(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logf(InfoLevel, format, v...)
	}
}

// Warn logs a message at WarnLevel, spaces are added between operands when neither is a string and a newline is appended.
func Warn(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Log(WarnLevel, v...)
	}
}

// Warnln logs a message at WarnLevel, spaces are always added between operands and a newline is appended.
func Warnln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logln(WarnLevel, v...)
	}
}

// Warnf logs a formatted message at WarnLevel.
func Warnf(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logf(WarnLevel, format, v...)
	}
}

// Error logs a message at ErrorLevel, spaces are added between operands when neither is a string and a newline is appended.
func Error(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Log(ErrorLevel, v...)
	}
}

// Errorln logs a message at ErrorLevel, spaces are always added between operands and a newline is appended.
func Errorln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logln(ErrorLevel, v...)
	}
}

// Errorf logs a formatted message at ErrorLevel.
func Errorf(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logf(ErrorLevel, format, v...)
	}
}

// DPanic logs a message at DPanicLevel, spaces are added between operands when neither is a string and a newline is appended.
func DPanic(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Log(DPanicLevel, v...)
	} else {
		panic(errors.New(fmt.Sprint(v...)))
	}
}

// DPanicln logs a message at DPanicLevel, spaces are always added between operands and a newline is appended.
func DPanicln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logln(DPanicLevel, v...)
	} else {
		panic(errors.New(fmt.Sprintln(v...)))
	}
}

// DPanicf logs a formatted message at DPanicLevel.
func DPanicf(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logf(DPanicLevel, format, v...)
	} else {
		panic(fmt.Errorf(format, v...))
	}
}

// Panic logs a message at PanicLevel, spaces are added between operands when neither is a string and a newline is appended.
func Panic(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Log(PanicLevel, v...)
	} else {
		panic(errors.New(fmt.Sprint(v...)))
	}
}

// Panicln logs a message at PanicLevel, spaces are always added between operands and a newline is appended.
func Panicln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logln(PanicLevel, v...)
	} else {
		panic(errors.New(fmt.Sprintln(v...)))
	}
}

// Panicf logs a formatted message at PanicLevel.
func Panicf(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logf(PanicLevel, format, v...)
	} else {
		panic(fmt.Errorf(format, v...))
	}
}

// Fatal logs a message at FatalLevel, spaces are added between operands when neither is a string and a newline is appended.
func Fatal(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Log(FatalLevel, v...)
	} else {
		os.Exit(1)
	}
}

// Fatalln logs a message at FatalLevel, spaces are always added between operands and a newline is appended.
func Fatalln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logln(FatalLevel, v...)
	} else {
		os.Exit(1)
	}
}

// Fatalf logs a formatted message at FatalLevel.
func Fatalf(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	if log, err := gxplugin.Using[Logger](pluginProvider, plugin.Name); err == nil {
		log.Logf(FatalLevel, format, v...)
	} else {
		os.Exit(1)
	}
}
