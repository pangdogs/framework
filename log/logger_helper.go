package log

import (
	gxplugin "kit.golaxy.org/golaxy/plugin"
)

// Trace logs a message at TraceLevel, spaces are added between operands when neither is a string and a newline is appended.
func Trace(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Log(TraceLevel, v...)
}

// Traceln logs a message at TraceLevel, spaces are always added between operands and a newline is appended.
func Traceln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Logln(TraceLevel, v...)
}

// Tracef logs a formatted message at TraceLevel.
func Tracef(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	Using(pluginProvider).Logf(TraceLevel, format, v...)
}

// Debug logs a message at DebugLevel, spaces are added between operands when neither is a string and a newline is appended.
func Debug(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Log(DebugLevel, v...)
}

// Debugln logs a message at DebugLevel, spaces are always added between operands and a newline is appended.
func Debugln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Logln(DebugLevel, v...)
}

// Debugf logs a formatted message at DebugLevel.
func Debugf(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	Using(pluginProvider).Logf(DebugLevel, format, v...)
}

// Info logs a message at InfoLevel, spaces are added between operands when neither is a string and a newline is appended.
func Info(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Log(InfoLevel, v...)
}

// Infoln logs a message at InfoLevel, spaces are always added between operands and a newline is appended.
func Infoln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Logln(InfoLevel, v...)
}

// Infof logs a formatted message at InfoLevel.
func Infof(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	Using(pluginProvider).Logf(InfoLevel, format, v...)
}

// Warn logs a message at WarnLevel, spaces are added between operands when neither is a string and a newline is appended.
func Warn(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Log(WarnLevel, v...)
}

// Warnln logs a message at WarnLevel, spaces are always added between operands and a newline is appended.
func Warnln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Logln(WarnLevel, v...)
}

// Warnf logs a formatted message at WarnLevel.
func Warnf(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	Using(pluginProvider).Logf(WarnLevel, format, v...)
}

// Error logs a message at ErrorLevel, spaces are added between operands when neither is a string and a newline is appended.
func Error(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Log(ErrorLevel, v...)
}

// Errorln logs a message at ErrorLevel, spaces are always added between operands and a newline is appended.
func Errorln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Logln(ErrorLevel, v...)
}

// Errorf logs a formatted message at ErrorLevel.
func Errorf(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	Using(pluginProvider).Logf(ErrorLevel, format, v...)
}

// DPanic logs a message at DPanicLevel, spaces are added between operands when neither is a string and a newline is appended.
func DPanic(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Log(DPanicLevel, v...)
}

// DPanicln logs a message at DPanicLevel, spaces are always added between operands and a newline is appended.
func DPanicln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Logln(DPanicLevel, v...)
}

// DPanicf logs a formatted message at DPanicLevel.
func DPanicf(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	Using(pluginProvider).Logf(DPanicLevel, format, v...)
}

// Panic logs a message at PanicLevel, spaces are added between operands when neither is a string and a newline is appended.
func Panic(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Log(PanicLevel, v...)
}

// Panicln logs a message at PanicLevel, spaces are always added between operands and a newline is appended.
func Panicln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Logln(PanicLevel, v...)
}

// Panicf logs a formatted message at PanicLevel.
func Panicf(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	Using(pluginProvider).Logf(PanicLevel, format, v...)
}

// Fatal logs a message at FatalLevel, spaces are added between operands when neither is a string and a newline is appended.
func Fatal(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Log(FatalLevel, v...)
}

// Fatalln logs a message at FatalLevel, spaces are always added between operands and a newline is appended.
func Fatalln(pluginProvider gxplugin.PluginProvider, v ...interface{}) {
	Using(pluginProvider).Logln(FatalLevel, v...)
}

// Fatalf logs a formatted message at FatalLevel.
func Fatalf(pluginProvider gxplugin.PluginProvider, format string, v ...interface{}) {
	Using(pluginProvider).Logf(FatalLevel, format, v...)
}
