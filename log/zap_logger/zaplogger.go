package zap_logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

// NewConsoleZapLogger 创建控制台样式zap logger（支持文件分割）
func NewConsoleZapLogger(level zapcore.Level, separator, fileName string, maxSize int, stdout, development bool) (*zap.Logger, zap.AtomicLevel) {
	var write zapcore.WriteSyncer

	if fileName != "" {
		rollingWrite := lumberjack.Logger{
			Filename: fileName,
			MaxSize:  maxSize,
		}
		write = zapcore.AddSync(&rollingWrite)
	}

	if stdout {
		if write != nil {
			write = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), write)
		} else {
			write = zapcore.AddSync(os.Stdout)
		}
	}

	if write == nil {
		panic("require at least one logger writer")
	}

	// 日志级别设置器
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(level)

	// 日志编码配置器
	encoderConfig := zapcore.EncoderConfig{
		LevelKey:         "level",
		TimeKey:          "timestamp",
		CallerKey:        "caller",
		StacktraceKey:    "stacktrace",
		MessageKey:       "msg",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.LowercaseLevelEncoder,
		EncodeTime:       zapcore.ISO8601TimeEncoder,
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: separator,
	}

	// 创建日志
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		write,
		atomicLevel,
	)

	options := []zap.Option{zap.AddCaller(), zap.AddStacktrace(zap.DPanicLevel)}
	if development {
		options = append(options, zap.Development())
	}

	logger := zap.New(core, options...)

	return logger, atomicLevel
}

// NewJsonZapLogger 创建Json样式zap logger（支持文件分割）
func NewJsonZapLogger(level zapcore.Level, fileName string, maxSize int, stdout, development bool) (*zap.Logger, zap.AtomicLevel) {
	var write zapcore.WriteSyncer

	if fileName != "" {
		rollingWrite := lumberjack.Logger{
			Filename: fileName,
			MaxSize:  maxSize,
		}
		write = zapcore.AddSync(&rollingWrite)
	}

	if stdout {
		if write != nil {
			write = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), write)
		} else {
			write = zapcore.AddSync(os.Stdout)
		}
	}

	if write == nil {
		panic("require at least one logger writer")
	}

	// 日志级别设置器
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(level)

	// 日志编码配置器
	encoderConfig := zapcore.EncoderConfig{
		LevelKey:       "level",
		TimeKey:        "timestamp",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建日志
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		write,
		atomicLevel,
	)

	options := []zap.Option{zap.AddCaller(), zap.AddStacktrace(zap.DPanicLevel)}
	if development {
		options = append(options, zap.Development())
	}

	logger := zap.New(core, options...)

	return logger, atomicLevel
}
