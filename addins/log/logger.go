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

// ILogger 提供结构化及糖化的 Zap 日志器。
type ILogger interface {
	// Logger 返回带当前 service 或 runtime 字段的结构化日志器。
	Logger() *zap.Logger
	// SugaredLogger 返回与 Logger 共享底层 Core 的 SugaredLogger。
	SugaredLogger() *zap.SugaredLogger
}

// L 返回 provider 上安装的结构化日志器；日志 add-in 未安装时会 panic。
func L(provider extension.AddInProvider) *zap.Logger {
	return AddIn.Require(provider).Logger()
}

// S 返回 provider 上安装的 SugaredLogger；日志 add-in 未安装时会 panic。
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

// JSON 创建一个在日志编码时才执行 json.Marshal 的字段。
// 编码失败时字段值会替换为描述错误的 JSON 字符串。
func JSON(key string, v any) zap.Field {
	return zap.Reflect(key, lazyJSON{v: v})
}

type lazyJSONRawStringer struct {
	v fmt.Stringer
}

func (l lazyJSONRawStringer) MarshalJSON() ([]byte, error) {
	if l.v == nil {
		return []byte("nil"), nil
	}
	return types.String2Bytes(l.v.String()), nil
}

// JSONRawStringer 创建一个延迟调用 String 的原始 JSON 字段。
// String 的返回值必须是有效 JSON。
func JSONRawStringer(key string, v fmt.Stringer) zap.Field {
	return zap.Reflect(key, lazyJSONRawStringer{v: v})
}

// JSONRawString 将 v 作为未经校验的原始 JSON 写入字段。
func JSONRawString(key string, v string) zap.Field {
	return zap.Any(key, json.RawMessage(types.String2Bytes(v)))
}

// JSONRawByteString 将 v 作为未经校验的原始 JSON 写入字段。
func JSONRawByteString(key string, v []byte) zap.Field {
	return zap.Any(key, json.RawMessage(v))
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

// Init 采用配置的 Zap 日志器（未提供时创建开发日志器），并附加当前 service 或 runtime 字段。
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

	var fields []zap.Field

	if rtCtx != nil {
		fields = append(fields, zap.Any("runtime", json.RawMessage(types.String2Bytes(rtCtx.String()))))
	} else {
		fields = append(fields, zap.Any("service", json.RawMessage(types.String2Bytes(svcCtx.String()))))
	}

	l.logger = logger.With(fields...)
	l.sugaredLogger = l.logger.Sugar()

	l.logger.Info("initializing add-in", zap.String("name", AddIn.Name))
}

// Shut 记录 add-in 停止并调用 Sync 刷新日志器；同步错误会被忽略。
func (l *_Logger) Shut(svcCtx service.Context, rtCtx runtime.Context) {
	l.logger.Info("shutting down add-in", zap.String("name", AddIn.Name))

	l.logger.Sync()
}

// RetainAfterTermination 使日志 add-in 在 Service 终止后继续可用。
// Runtime 不处理此标记，安装到 Runtime 的日志 add-in 仍会正常 Shut。
func (*_Logger) RetainAfterTermination() {}

// Logger 返回附加了当前 service 或 runtime 字段的结构化日志器。
func (l *_Logger) Logger() *zap.Logger {
	return l.logger
}

// SugaredLogger 返回与 Logger 共享底层 Core 的 SugaredLogger。
func (l *_Logger) SugaredLogger() *zap.SugaredLogger {
	return l.sugaredLogger
}
