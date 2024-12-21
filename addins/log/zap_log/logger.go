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

package zap_log

import (
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/log"
	"go.uber.org/zap"
)

func newLogger(settings ...option.Setting[LoggerOptions]) log.ILogger {
	return &_Logger{
		options: option.Make(With.Default(), settings...),
	}
}

type _Logger struct {
	options       LoggerOptions
	sugaredLogger *zap.SugaredLogger
}

// Init 初始化插件
func (l *_Logger) Init(svcCtx service.Context, rtCtx runtime.Context) {
	options := []zap.Option{zap.AddCallerSkip(l.options.CallerSkip)}
	if l.options.ServiceInfo {
		options = append(options, zap.Fields(zap.String("service", svcCtx.String())))
	}

	if rtCtx != nil {
		if l.options.RuntimeInfo {
			options = append(options, zap.Fields(zap.String("runtime", rtCtx.String())))
		}
	}

	l.sugaredLogger = l.options.ZapLogger.WithOptions(options...).Sugar()
}

// Log writes a log entry, spaces are added between operands when neither is a string and a newline is appended
func (l *_Logger) Log(level log.Level, v ...interface{}) {
	sugaredLogger := l.sugaredLogger

	switch level {
	case log.TraceLevel:
		sugaredLogger.Debug(v...)
	case log.DebugLevel:
		sugaredLogger.Debug(v...)
	case log.InfoLevel:
		sugaredLogger.Info(v...)
	case log.WarnLevel:
		sugaredLogger.Warn(v...)
	case log.ErrorLevel:
		sugaredLogger.Error(v...)
	case log.DPanicLevel:
		sugaredLogger.DPanic(v...)
	case log.PanicLevel:
		sugaredLogger.Panic(v...)
	case log.FatalLevel:
		sugaredLogger.Fatal(v...)
	}
}

// Logln writes a log entry, spaces are always added between operands and a newline is appended
func (l *_Logger) Logln(level log.Level, v ...interface{}) {
	sugaredLogger := l.sugaredLogger

	switch level {
	case log.TraceLevel:
		sugaredLogger.Debugln(v...)
	case log.DebugLevel:
		sugaredLogger.Debugln(v...)
	case log.InfoLevel:
		sugaredLogger.Infoln(v...)
	case log.WarnLevel:
		sugaredLogger.Warnln(v...)
	case log.ErrorLevel:
		sugaredLogger.Errorln(v...)
	case log.DPanicLevel:
		sugaredLogger.DPanicln(v...)
	case log.PanicLevel:
		sugaredLogger.Panicln(v...)
	case log.FatalLevel:
		sugaredLogger.Fatalln(v...)
	}
}

// Logf writes a formatted log entry
func (l *_Logger) Logf(level log.Level, format string, v ...interface{}) {
	sugaredLogger := l.sugaredLogger

	switch level {
	case log.TraceLevel:
		sugaredLogger.Debugf(format, v...)
	case log.DebugLevel:
		sugaredLogger.Debugf(format, v...)
	case log.InfoLevel:
		sugaredLogger.Infof(format, v...)
	case log.WarnLevel:
		sugaredLogger.Warnf(format, v...)
	case log.ErrorLevel:
		sugaredLogger.Errorf(format, v...)
	case log.DPanicLevel:
		sugaredLogger.DPanicf(format, v...)
	case log.PanicLevel:
		sugaredLogger.Panicf(format, v...)
	case log.FatalLevel:
		sugaredLogger.Fatalf(format, v...)
	}
}
