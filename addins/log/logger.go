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
	"encoding/json"
	"fmt"
	"os"

	"git.golaxy.org/core/extension"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ILogger interface {
	Logger() *zap.Logger
	SugaredLogger() *zap.SugaredLogger
}

func L(provider extension.AddInProvider) *zap.Logger {
	return AddIn.Require(provider).Logger()
}

func S(provider extension.AddInProvider) *zap.SugaredLogger {
	return AddIn.Require(provider).SugaredLogger()
}

type lazyJSON struct {
	v any
}

func (l lazyJSON) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(l.v)
	if err != nil {
		return json.Marshal(fmt.Sprintf("json.Marshal(): %s", err.Error()))
	}
	return data, nil
}

func JSON(key string, v any) zap.Field {
	return zap.Reflect(key, lazyJSON{v: v})
}

func newLogger(settings ...option.Setting[LoggerOptions]) ILogger {
	return &_Logger{
		options: option.New(With.Default(), settings...),
	}
}

type _Logger struct {
	options       LoggerOptions
	logger        *zap.Logger
	sugaredLogger *zap.SugaredLogger
}

// Init 初始化插件
func (l *_Logger) Init(svcCtx service.Context, rtCtx runtime.Context) {
	logger := l.options.Logger
	if logger == nil {
		logger = zap.New(
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
				zapcore.AddSync(os.Stdout),
				zapcore.DebugLevel,
			),
			zap.AddCaller(),
			zap.AddStacktrace(zap.DPanicLevel),
		)
	}

	fields := []zap.Field{zap.Any("service", json.RawMessage(types.String2Bytes(svcCtx.String())))}

	if rtCtx != nil {
		fields = append(fields, zap.Any("runtime", json.RawMessage(types.String2Bytes(rtCtx.String()))))
	}

	l.logger = logger.With(fields...)
	l.sugaredLogger = l.logger.Sugar()

	l.logger.Info("initializing add-in", zap.String("name", AddIn.Name))
}

// Shut 关闭插件
func (l *_Logger) Shut(svcCtx service.Context, rtCtx runtime.Context) {
	l.logger.Info("shutting down add-in", zap.String("name", AddIn.Name))
	
	l.logger.Sync()
}

func (l *_Logger) Logger() *zap.Logger {
	return l.logger
}

func (l *_Logger) SugaredLogger() *zap.SugaredLogger {
	return l.sugaredLogger
}
