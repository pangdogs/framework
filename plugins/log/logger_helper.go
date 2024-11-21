/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package log

import (
	"git.golaxy.org/core/extension"
)

// Trace logs a message at TraceLevel, spaces are added between operands when neither is a string and a newline is appended.
func Trace(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Log(TraceLevel, v...)
}

// Traceln logs a message at TraceLevel, spaces are always added between operands and a newline is appended.
func Traceln(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Logln(TraceLevel, v...)
}

// Tracef logs a formatted message at TraceLevel.
func Tracef(provider extension.PluginProvider, format string, v ...interface{}) {
	Using(provider).Logf(TraceLevel, format, v...)
}

// Debug logs a message at DebugLevel, spaces are added between operands when neither is a string and a newline is appended.
func Debug(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Log(DebugLevel, v...)
}

// Debugln logs a message at DebugLevel, spaces are always added between operands and a newline is appended.
func Debugln(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Logln(DebugLevel, v...)
}

// Debugf logs a formatted message at DebugLevel.
func Debugf(provider extension.PluginProvider, format string, v ...interface{}) {
	Using(provider).Logf(DebugLevel, format, v...)
}

// Info logs a message at InfoLevel, spaces are added between operands when neither is a string and a newline is appended.
func Info(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Log(InfoLevel, v...)
}

// Infoln logs a message at InfoLevel, spaces are always added between operands and a newline is appended.
func Infoln(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Logln(InfoLevel, v...)
}

// Infof logs a formatted message at InfoLevel.
func Infof(provider extension.PluginProvider, format string, v ...interface{}) {
	Using(provider).Logf(InfoLevel, format, v...)
}

// Warn logs a message at WarnLevel, spaces are added between operands when neither is a string and a newline is appended.
func Warn(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Log(WarnLevel, v...)
}

// Warnln logs a message at WarnLevel, spaces are always added between operands and a newline is appended.
func Warnln(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Logln(WarnLevel, v...)
}

// Warnf logs a formatted message at WarnLevel.
func Warnf(provider extension.PluginProvider, format string, v ...interface{}) {
	Using(provider).Logf(WarnLevel, format, v...)
}

// Error logs a message at ErrorLevel, spaces are added between operands when neither is a string and a newline is appended.
func Error(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Log(ErrorLevel, v...)
}

// Errorln logs a message at ErrorLevel, spaces are always added between operands and a newline is appended.
func Errorln(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Logln(ErrorLevel, v...)
}

// Errorf logs a formatted message at ErrorLevel.
func Errorf(provider extension.PluginProvider, format string, v ...interface{}) {
	Using(provider).Logf(ErrorLevel, format, v...)
}

// DPanic logs a message at DPanicLevel, spaces are added between operands when neither is a string and a newline is appended.
func DPanic(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Log(DPanicLevel, v...)
}

// DPanicln logs a message at DPanicLevel, spaces are always added between operands and a newline is appended.
func DPanicln(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Logln(DPanicLevel, v...)
}

// DPanicf logs a formatted message at DPanicLevel.
func DPanicf(provider extension.PluginProvider, format string, v ...interface{}) {
	Using(provider).Logf(DPanicLevel, format, v...)
}

// Panic logs a message at PanicLevel, spaces are added between operands when neither is a string and a newline is appended.
func Panic(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Log(PanicLevel, v...)
}

// Panicln logs a message at PanicLevel, spaces are always added between operands and a newline is appended.
func Panicln(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Logln(PanicLevel, v...)
}

// Panicf logs a formatted message at PanicLevel.
func Panicf(provider extension.PluginProvider, format string, v ...interface{}) {
	Using(provider).Logf(PanicLevel, format, v...)
}

// Fatal logs a message at FatalLevel, spaces are added between operands when neither is a string and a newline is appended.
func Fatal(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Log(FatalLevel, v...)
}

// Fatalln logs a message at FatalLevel, spaces are always added between operands and a newline is appended.
func Fatalln(provider extension.PluginProvider, v ...interface{}) {
	Using(provider).Logln(FatalLevel, v...)
}

// Fatalf logs a formatted message at FatalLevel.
func Fatalf(provider extension.PluginProvider, format string, v ...interface{}) {
	Using(provider).Logf(FatalLevel, format, v...)
}
